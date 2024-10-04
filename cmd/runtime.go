package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type GossResult struct {
	Results []struct {
		Duration     int64       `json:"duration"`
		Err          interface{} `json:"err"`
		Property     string      `json:"property"`
		ResourceID   string      `json:"resource-id"`
		ResourceType string      `json:"resource-type"`
		Result       int         `json:"result"`
		Skipped      bool        `json:"skipped"`
		Successful   bool        `json:"successful"`
		SummaryLine  string      `json:"summary-line"`
		Title        string      `json:"title"`
	} `json:"results"`
	Summary struct {
		FailedCount   int    `json:"failed-count"`
		SkippedCount  int    `json:"skipped-count"`
		SummaryLine   string `json:"summary-line"`
		TestCount     int    `json:"test-count"`
		TotalDuration int64  `json:"total-duration"`
	} `json:"summary"`
}

func ProcessRuntimeType(policy Policy, gossPath string, targetDir string, filePaths []string, isObserve bool) error {

	if policy.Type != "runtime" {
		return fmt.Errorf("invalid policy type for runtime processing: %s", policy.Type)
	}

	gossConfigPath := policy.Runtime.Config
	if gossConfigPath == "" {
		return fmt.Errorf("goss runtime config path is not specified in the policy")
	}

	// Handle relative paths
	if !filepath.IsAbs(gossConfigPath) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Error().Err(err).Msg("error getting current working directory")
			return fmt.Errorf("error getting current working directory: %w", err)
		}
		gossConfigPath = filepath.Join(cwd, gossConfigPath)
	}

	// Ensure the goss config file exists
	if _, err := os.Stat(gossConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("goss config file does not exist: %s", gossConfigPath)
	}

	//log.Debug().Msgf("Using goss config file: %s ", gossConfigPath)
	var args []string

	// Prepare the goss validate command
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		args = append(args, "--use-alpha=1")
	}

	args = append(args, "-g", gossConfigPath, "validate", "--format", "json")

	log.Debug().Msgf("Args: %s", args)

	// Run goss validate
	cmd := exec.Command(gossPath, args...)
	output, _ := cmd.CombinedOutput()

	// log.Debug().Msgf(" Output: %s", string(output))

	var gossResult GossResult
	if err := json.Unmarshal(output, &gossResult); err != nil {
		return fmt.Errorf("error parsing goss output: %w Output: %s", err, string(output))
	}

	// Generate SARIF report
	sarifReport, err := generateRuntimeSARIFReport(policy, gossResult)
	if err != nil {
		log.Error().Err(err).Msg("error generating SARIF report")
		return fmt.Errorf("error generating SARIF report: %w", err)
	}

	// Write SARIF report to file
	var sarifOutputFile string

	if policy.RunID != "" {
		if err := writeSARIFReport(policy.RunID, sarifReport); err != nil {
			log.Error().Err(err).Msg("error writing SARIF report")
			return fmt.Errorf("error writing SARIF report: %w", err)
		}
		sarifOutputFile = fmt.Sprintf("%s.sarif", policy.RunID)
	} else {
		if err := writeSARIFReport(policy.ID, sarifReport); err != nil {
			log.Error().Err(err).Msg("error writing SARIF report")
			return fmt.Errorf("error writing SARIF report: %w", err)
		}
		sarifOutputFile = fmt.Sprintf("%s.sarif", NormalizeFilename(policy.ID))

	}
	if isObserve {

		if len(sarifReport.Runs) == 0 {
			log.Warn().Msg("Runtime SARIF contains no runs")
		} else {
			// Post merged SARIF report to webhooks
			if err := PostResultsToWebhooks(sarifReport); err != nil {
				log.Error().Err(err).Msg("Failed to post Runtime Results to webhooks")
			}
		}

	}

	log.Debug().Msgf("Runtime policy %s processed. SARIF report written to: %s ", policy.ID, sarifOutputFile)

	return nil
}

func generateRuntimeSARIFReport(policy Policy, gossResult GossResult) (SARIFReport, error) {

	timestamp := time.Now().Format(time.RFC3339)

	sarifReport := SARIFReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						FullName:        "INTERCEPT",
						Name:            "INTERCEPT",
						Version:         buildVersion,
						SemanticVersion: buildVersion,
					},
				},
				Results: []Result{},
				Invocations: []Invocation{
					{
						ExecutionSuccessful: true,
						Properties: InvocationProperties{
							ReportCompliant: true,
						},
					},
				},
			},
		},
	}

	policyLevel := getPolicyLevel(policy)

	if policy.RunID == "" {
		policy.RunID = "N/A"
	}

	for _, res := range gossResult.Results {
		var sarifLevel SARIFLevel
		var messageText string

		if res.Successful && !res.Skipped {
			sarifLevel = SARIFNote
		} else if res.Successful && res.Skipped {
			sarifLevel = SARIFWarning
		} else {
			sarifLevel = policyLevel
		}

		if res.Err != nil {
			messageText = fmt.Sprintf("Error: %v", res.Err)
		} else {
			messageText = res.SummaryLine
		}

		sarifResult := Result{
			RuleID:  policy.ID,
			Level:   sarifLevel,
			Message: Message{Text: messageText},
			Locations: []Location{
				{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: ArtifactLocation{URI: res.ResourceID},
					},
				},
			},
			Properties: ResultProperties{
				ResourceType:    res.ResourceType,
				Property:        res.Property,
				ResultType:      "detail",
				ObserveRunId:    policy.RunID,
				ResultTimestamp: timestamp,
				Environment:     environment,
				Name:            policy.Metadata.Name,
				Description:     policy.Metadata.Description,
				MsgError:        policy.Metadata.MsgError,
				MsgSolution:     policy.Metadata.MsgSolution,
				SarifInt:        sarifLevelToInt(sarifLevel),
			},
		}
		sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, sarifResult)
	}

	summaryLevel := getSummaryLevel(gossResult.Summary, policyLevel)

	// Add overall summary as a separate result
	summarySarifResult := Result{
		RuleID:  policy.ID,
		Level:   summaryLevel,
		Message: Message{Text: gossResult.Summary.SummaryLine},
		Locations: []Location{
			{
				PhysicalLocation: PhysicalLocation{
					ArtifactLocation: ArtifactLocation{URI: "N/A"},
				},
			},
		},
		Properties: ResultProperties{
			ResultType:      "summary",
			ObserveRunId:    policy.RunID,
			ResultTimestamp: timestamp,
			Environment:     environment,
			Name:            policy.Metadata.Name,
			Description:     policy.Metadata.Description,
			MsgError:        policy.Metadata.MsgError,
			MsgSolution:     policy.Metadata.MsgSolution,
			SarifInt:        sarifLevelToInt(summaryLevel),
		},
	}
	sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, summarySarifResult)

	sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant = summaryLevel == SARIFNote

	if outputTypeMatrixConfig.LOG {
		PostResultsToComplianceLog(sarifReport)
	}

	return sarifReport, nil
}

func getPolicyLevel(policy Policy) SARIFLevel {
	fatal := policy.Enforcement[0].Fatal == "true"
	if fatal {
		return SARIFError
	}
	return SARIFWarning
}

func getSummaryLevel(summary struct {
	FailedCount   int    `json:"failed-count"`
	SkippedCount  int    `json:"skipped-count"`
	SummaryLine   string `json:"summary-line"`
	TestCount     int    `json:"test-count"`
	TotalDuration int64  `json:"total-duration"`
}, policyLevel SARIFLevel) SARIFLevel {
	if summary.FailedCount > 0 {
		return policyLevel
	}
	return SARIFNote
}
