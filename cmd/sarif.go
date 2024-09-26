package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charlievieth/fastwalk"
)

// SARIFLevel represents the severity level in SARIF format
type SARIFLevel string

var sarifDir string
var mergeOutputPath string

const (
	SARIFError   SARIFLevel = "error"
	SARIFWarning SARIFLevel = "warning"
	SARIFNote    SARIFLevel = "note"
	SARIFNone    SARIFLevel = "none"
)

// PolicyToSARIFLevel translates policy enforcement settings to SARIF level
func PolicyToSARIFLevel(fatal bool, exceptions bool, confidence string) SARIFLevel {
	switch {
	case fatal && !exceptions:
		return SARIFError
	case fatal && exceptions:
		return SARIFError
	case !fatal && !exceptions && strings.EqualFold(confidence, "high"):
		return SARIFError
	case !fatal && exceptions && strings.EqualFold(confidence, "high"):
		return SARIFWarning
	case !fatal && !exceptions && strings.EqualFold(confidence, "low"):
		return SARIFWarning
	case !fatal && exceptions && strings.EqualFold(confidence, "low"):
		return SARIFNote
	case !fatal && exceptions && strings.EqualFold(confidence, "info"):
		return SARIFNone
	default:
		return SARIFWarning
	}
}

func sarifLevelToString(level SARIFLevel) string {
	switch level {
	case SARIFNone:
		return "none"
	case SARIFNote:
		return "note"
	case SARIFWarning:
		return "warning"
	case SARIFError:
		return "error"
	default:
		return "unknown"
	}
}

func selectEnforcementRule(policy Policy, environment string) Enforcement {
	for _, rule := range policy.Enforcement {
		if rule.Environment == environment || rule.Environment == "all" {
			return rule
		}
	}
	// If no matching rule is found, return a default rule or the first rule
	if len(policy.Enforcement) > 0 {
		return policy.Enforcement[0]
	}
	// If there are no rules at all, return a default rule
	return Enforcement{
		Environment: "all",
		Fatal:       "false",
		Exceptions:  "false",
		Confidence:  "low",
	}
}

func calculateSARIFLevel(policy Policy, environment string) SARIFLevel {
	selectedRule := selectEnforcementRule(policy, environment)
	return PolicyToSARIFLevel(
		selectedRule.Fatal == "true",
		selectedRule.Exceptions == "true",
		selectedRule.Confidence,
	)
}

// RipgrepOutput represents the structure of the ripgrep JSON output
type RipgrepOutput struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		LineNumber int `json:"line_number"`
		Lines      struct {
			Text string `json:"text"`
		} `json:"lines"`
		Submatches []struct {
			Match struct {
				Text string `json:"text"`
			} `json:"match"`
		} `json:"submatches"`
	} `json:"data"`
}

// SARIFReport represents the structure of a SARIF report
type SARIFReport struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []Run  `json:"runs"`
}

type Run struct {
	Tool        Tool         `json:"tool"`
	Results     []Result     `json:"results"`
	Invocations []Invocation `json:"invocations"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Result struct {
	RuleID     string            `json:"ruleId"`
	Level      SARIFLevel        `json:"level"`
	Message    Message           `json:"message"`
	Locations  []Location        `json:"locations"`
	Properties map[string]string `json:"properties,omitempty"`
}

type Message struct {
	Text string `json:"text"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region"`
}

type ArtifactLocation struct {
	URI string `json:"uri"`
}

type Region struct {
	StartLine   int     `json:"startLine"`
	StartColumn int     `json:"startColumn"`
	EndColumn   int     `json:"endColumn"`
	Snippet     Snippet `json:"snippet"`
}

type Snippet struct {
	Text string `json:"text"`
}

type Invocation struct {
	ExecutionSuccessful bool              `json:"executionSuccessful"`
	CommandLine         string            `json:"commandLine,omitempty"`
	Properties          map[string]string `json:"properties,omitempty"`
}

type Notification struct {
	Descriptor Descriptor `json:"descriptor"`
	Message    Message    `json:"message"`
}

type Descriptor struct {
	ID string `json:"id"`
}

func GenerateSARIFReport(inputFile string, policy Policy) (SARIFReport, error) {
	var sarifReport SARIFReport
	timestamp := time.Now().Format(time.RFC3339)

	// Read and parse the input JSON file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return sarifReport, fmt.Errorf("error reading input file: %w", err)
	}

	var rgOutputs []RipgrepOutput
	err = json.Unmarshal(inputData, &rgOutputs)
	if err != nil {
		return sarifReport, fmt.Errorf("error unmarshaling input JSON: %w", err)
	}

	// Create SARIF report structure
	sarifReport = SARIFReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:    "Intercept",
						Version: buildVersion,
					},
				},
				Results: []Result{},
				Invocations: []Invocation{
					{
						ExecutionSuccessful: true,
					},
				},
			},
		},
	}

	var results []Result
	noMatches := len(rgOutputs) == 1 && rgOutputs[0].Type == "summary"

	if policy.Type == "scan" && noMatches {
		// For SCAN policies with no matches, add a compliance result
		results = append(results, Result{
			RuleID: policy.ID,
			Level:  SARIFNote,
			Message: Message{
				Text: fmt.Sprintf("Policy %s is compliant: No violations found", policy.ID),
			},
			Locations: []Location{
				{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: ArtifactLocation{URI: "N/A"},
					},
				},
			},
			Properties: map[string]string{
				"result-type":      "detail",
				"observe-run-id":   policy.RunID,
				"result-timestamp": timestamp,
				"name":             policy.Metadata.Name,
				"description":      policy.Metadata.Description,
				"msg-error":        policy.Metadata.MsgError,
				"msg-solution":     policy.Metadata.MsgSolution,
				"note":             "true",
			},
		})
	} else {
		// Process ripgrep output and add results to SARIF report
		for _, rgOutput := range rgOutputs {
			if rgOutput.Type == "match" {

				sarifLevel := calculateSARIFLevel(policy, environment)
				levelProperty := sarifLevelToString(sarifLevel)
				for _, submatch := range rgOutput.Data.Submatches {
					matchText := submatch.Match.Text
					startColumn := strings.Index(rgOutput.Data.Lines.Text, matchText) + 1
					endColumn := startColumn + len(matchText)

					result := Result{
						RuleID: policy.ID,
						Level:  sarifLevel,
						Message: Message{
							Text: fmt.Sprintf("Policy violation: %s Matched text: %s", policy.Metadata.Name, matchText),
						},
						Locations: []Location{
							{
								PhysicalLocation: PhysicalLocation{
									ArtifactLocation: ArtifactLocation{URI: rgOutput.Data.Path.Text},
									Region: Region{
										StartLine:   rgOutput.Data.LineNumber,
										StartColumn: startColumn,
										EndColumn:   endColumn,
										Snippet: Snippet{
											Text: matchText,
										},
									},
								},
							},
						},
						Properties: map[string]string{
							"result-type":      "detail",
							"observe-run-id":   policy.RunID,
							"result-timestamp": timestamp,
							"name":             policy.Metadata.Name,
							"description":      policy.Metadata.Description,
							"msg-error":        policy.Metadata.MsgError,
							"msg-solution":     policy.Metadata.MsgSolution,
							levelProperty:      "true",
						},
					}
					results = append(results, result)
				}
			}
		}
	}

	sarifReport.Runs[0].Results = results
	return sarifReport, nil
}

// Updated to return SARIFReport instead of writing to a file
func GenerateAssureSARIFReport(inputFile string, policy Policy, status string) (SARIFReport, error) {
	var sarifReport SARIFReport

	// Read and parse the input JSON file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return sarifReport, fmt.Errorf("error reading input file: %w", err)
	}

	var rgOutputs []RipgrepOutput
	err = json.Unmarshal(inputData, &rgOutputs)
	if err != nil {
		return sarifReport, fmt.Errorf("error unmarshaling input JSON: %w", err)
	}

	// Create SARIF report structure
	sarifReport = SARIFReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:    "Intercept",
						Version: buildVersion,
					},
				},
				Results: []Result{},
				Invocations: []Invocation{
					{
						ExecutionSuccessful: true,
					},
				},
			},
		},
	}

	// Translate policy enforcement to SARIF level

	sarifLevel := calculateSARIFLevel(policy, environment)
	timestamp := time.Now().Format(time.RFC3339)
	levelProperty := sarifLevelToString(sarifLevel)

	result := Result{}
	if status == "FOUND" {
		// For assure policies, we report based on the status
		result = Result{
			RuleID: policy.ID,
			Level:  "note",
			Message: Message{
				Text: fmt.Sprintf("Assure policy %s: Pattern %s", policy.Metadata.Name, status),
			},
			Properties: map[string]string{
				"result-type":      "detail",
				"observe-run-id":   policy.RunID,
				"result-timestamp": timestamp,
				"name":             policy.Metadata.Name,
				"description":      policy.Metadata.Description,
				"msg-error":        policy.Metadata.MsgError,
				"msg-solution":     policy.Metadata.MsgSolution,
				"note":             "true",
			},
		}
	} else {
		// For assure policies, we report based on the status
		result = Result{
			RuleID: policy.ID,
			Level:  sarifLevel,
			Message: Message{
				Text: fmt.Sprintf("Assure policy %s: Pattern %s", policy.Metadata.Name, status),
			},
			Properties: map[string]string{
				"result-type":      "detail",
				"observe-run-id":   policy.RunID,
				"result-timestamp": timestamp,
				"name":             policy.Metadata.Name,
				"description":      policy.Metadata.Description,
				"msg-error":        policy.Metadata.MsgError,
				"msg-solution":     policy.Metadata.MsgSolution,
				levelProperty:      "true",
			},
		}
	}

	if len(rgOutputs) > 0 && status == "FOUND" {
		// If matches were found, include location information
		for _, rgOutput := range rgOutputs {
			if rgOutput.Type == "match" {
				result.Locations = append(result.Locations, Location{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: ArtifactLocation{URI: rgOutput.Data.Path.Text},
						Region: Region{
							StartLine:   rgOutput.Data.LineNumber,
							StartColumn: 1,
							EndColumn:   len(rgOutput.Data.Lines.Text),
						},
					},
				})
			}
		}
	} else {
		// If no matches were found or status is "NOT FOUND", we don't have specific location information
		result.Locations = []Location{
			{
				PhysicalLocation: PhysicalLocation{
					ArtifactLocation: ArtifactLocation{URI: "N/A"},
					Region: Region{
						StartLine:   1,
						StartColumn: 1,
						EndColumn:   1,
					},
				},
			},
		}
	}

	sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, result)

	return sarifReport, nil
}

func GenerateSchemaSARIFReport(policy Policy, filePath string, valid bool, issues []string, patched bool) SARIFReport {
	sarifReport := SARIFReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:    "Intercept",
						Version: buildVersion,
					},
				},
				Results:     []Result{},
				Invocations: []Invocation{{ExecutionSuccessful: true}},
			},
		},
	}

	sarifLevel := calculateSARIFLevel(policy, environment)

	levelProperty := sarifLevelToString(sarifLevel)

	timestamp := time.Now().Format(time.RFC3339)

	if !valid {
		for _, issue := range issues {
			result := Result{
				RuleID: policy.ID,
				Level:  sarifLevel,
				Message: Message{
					Text: fmt.Sprintf("Validation issue: %s", issue),
				},
				Locations: []Location{
					{
						PhysicalLocation: PhysicalLocation{
							ArtifactLocation: ArtifactLocation{
								URI: filepath.ToSlash(filePath),
							},
						},
					},
				},
				Properties: map[string]string{
					"result-type":      "detail",
					"observe-run-id":   policy.RunID,
					"result-timestamp": timestamp,
					"name":             policy.Metadata.Name,
					"description":      policy.Metadata.Description,
					"msg-error":        policy.Metadata.MsgError,
					"msg-solution":     policy.Metadata.MsgSolution,
					levelProperty:      "true",
				},
			}
			sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, result)
		}
	}

	// Add a summary result
	summaryText := fmt.Sprintf("Validation %s for policy %s",
		map[bool]string{true: "passed", false: "failed"}[valid],
		policy.ID)
	if patched {
		summaryText += " (content patched)"
	}

	detailSarif := map[bool]SARIFLevel{true: SARIFNote, false: sarifLevel}[valid]
	levelProperty = sarifLevelToString(detailSarif)

	summaryResult := Result{
		RuleID: policy.ID,
		Level:  detailSarif,
		Message: Message{
			Text: summaryText,
		},
		Locations: []Location{
			{
				PhysicalLocation: PhysicalLocation{
					ArtifactLocation: ArtifactLocation{
						URI: filepath.ToSlash(filePath),
					},
				},
			},
		},
		Properties: map[string]string{
			"result-type":      "detail",
			"observe-run-id":   policy.RunID,
			"result-timestamp": timestamp,
			"name":             policy.Metadata.Name,
			"description":      policy.Metadata.Description,
			"msg-error":        policy.Metadata.MsgError,
			"msg-solution":     policy.Metadata.MsgSolution,
			levelProperty:      "true",
		},
	}
	sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, summaryResult)

	return sarifReport
}

func initSARIFProcessing() error {
	var err error
	sarifDir, err = os.MkdirTemp("", "sarif_reports")
	if err != nil {
		log.Error().Err(err).Msg("failed to create temporary directory for SARIF reports")
		return fmt.Errorf("failed to create temporary directory for SARIF reports: %w", err)
	}
	return nil
}

func writeSARIFReport(policyID string, report SARIFReport) error {
	filename := filepath.Join("_sarif", fmt.Sprintf("%s.sarif", NormalizeFilename(policyID)))
	if outputDir != "" {
		filename = filepath.Join(outputDir, filename)
	}
	file, err := os.Create(filename)
	if err != nil {
		log.Error().Err(err).Msg("failed to create SARIF file for policy %s")
		return fmt.Errorf("failed to create SARIF file for policy %s: %w", policyID, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		log.Error().Err(err).Msg("failed to write SARIF report for policy %s")
		return fmt.Errorf("failed to write SARIF report for policy %s: %w", policyID, err)
	}

	return nil
}
func MergeSARIFReports(commandLine string, perf Performance, isScheduled bool) (SARIFReport, error) {
	searchDir := "_sarif"
	if outputDir != "" {
		searchDir = filepath.Join(outputDir, searchDir)
	}

	files, err := filepath.Glob(filepath.Join(searchDir, "*.sarif"))
	if err != nil {
		log.Error().Err(err).Msg("failed to list SARIF files")
		return SARIFReport{}, fmt.Errorf("failed to list SARIF files: %w", err)
	}

	if len(files) == 0 {
		log.Warn().Msgf("no SARIF files found in %s directory, skipping merge", searchDir)
		return SARIFReport{}, nil
	}

	timestamp := time.Now().Format(time.RFC3339)

	mergedReport := SARIFReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:    "Intercept",
						Version: buildVersion,
					},
				},
				Results: []Result{},
				Invocations: []Invocation{
					{
						ExecutionSuccessful: true,
						CommandLine:         commandLine,
						Properties: map[string]string{
							"run-id":            intercept_run_id,
							"start-time":        perf.StartTime.Format(time.RFC3339),
							"end-time":          perf.EndTime.Format(time.RFC3339),
							"execution-time-ms": fmt.Sprintf("%d", perf.Delta.Milliseconds()),
							"environment":       environment,
							"debug":             fmt.Sprintf("%v", debugOutput),
							"report-timestamp":  timestamp,
							"host-data":         hostData,
							"host-fingerprint":  hostFingerprint,
						},
					},
				},
			},
		},
	}

	isCompliant := true

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Error().Err(err).Msg("failed to read SARIF file %s")
			return SARIFReport{}, fmt.Errorf("failed to read SARIF file %s: %w", file, err)
		}

		var report SARIFReport
		if err := json.Unmarshal(data, &report); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal SARIF report from %s")
			return SARIFReport{}, fmt.Errorf("failed to unmarshal SARIF report from %s: %w", file, err)
		}

		for _, run := range report.Runs {
			mergedReport.Runs[0].Results = append(mergedReport.Runs[0].Results, run.Results...)

			for _, result := range run.Results {
				if result.Level == SARIFWarning || result.Level == SARIFError {
					isCompliant = false
					break
				}
			}
		}
	}

	// Set the compliance status
	if isCompliant {
		mergedReport.Runs[0].Invocations[0].Properties["report-status"] = "compliant"
		mergedReport.Runs[0].Invocations[0].Properties["report-compliant"] = "true"
	} else {
		mergedReport.Runs[0].Invocations[0].Properties["report-status"] = "non-compliant"
		mergedReport.Runs[0].Invocations[0].Properties["report-compliant"] = "false"
	}

	if !isScheduled {
		mergeOutputPath = fmt.Sprintf("intercept_%s.sarif.json", intercept_run_id[:6])
	} else {
		timestamp := time.Now().UTC().Format("20060102T150405Z")
		mergeOutputPath = fmt.Sprintf("%s_intercept_%s.sarif.json", timestamp, intercept_run_id[:6])
		mergeOutputPath = filepath.Join("_status", mergeOutputPath)
	}

	if outputDir != "" {
		mergeOutputPath = filepath.Join(outputDir, mergeOutputPath)
	}

	mergedData, err := json.MarshalIndent(mergedReport, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal merged SARIF report")
		return SARIFReport{}, fmt.Errorf("failed to marshal merged SARIF report: %w", err)
	}

	if err := os.WriteFile(mergeOutputPath, mergedData, 0644); err != nil {
		log.Error().Err(err).Msg("failed to write merged SARIF report")
		return SARIFReport{}, fmt.Errorf("failed to write merged SARIF report: %w", err)
	}

	log.Debug().Msgf("SARIF Report written to: %s ", mergeOutputPath)

	return mergedReport, nil
}

func cleanupSARIFProcessing() {
	os.RemoveAll(sarifDir)
}

func cleanupSARIFFolder() error {

	sarifDir := "_sarif"
	if outputDir != "" {
		sarifDir = filepath.Join(outputDir, sarifDir)
	}

	err := fastwalk.Walk(nil, sarifDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // Error accessing the path
		}

		if !d.IsDir() || path != sarifDir { // Skip the root dir itself
			err := os.RemoveAll(path)
			if err != nil {
				return err // Stop on error
			}
		}
		return nil
	})

	return err
}

func createSARIFReport(results []Result) SARIFReport {
	return SARIFReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:    "Intercept",
						Version: buildVersion,
					},
				},
				Results:     results,
				Invocations: []Invocation{{ExecutionSuccessful: true}},
			},
		},
	}
}
func GenerateAPISARIFReport(policy Policy, endpoint string, matchFound bool, issues []string) (SARIFReport, error) {
	sarifReport := SARIFReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:    "Intercept",
						Version: buildVersion,
					},
				},
				Results:     []Result{},
				Invocations: []Invocation{{ExecutionSuccessful: true}},
			},
		},
	}

	sarifLevel := calculateSARIFLevel(policy, environment)

	var resultLevel SARIFLevel
	var resultText string

	if len(policy.Regex) > 0 {
		// For regex-based API checks (assure-like behavior)
		if matchFound {
			resultLevel = SARIFNote
			resultText = fmt.Sprintf("API assurance passed for policy %s: Pattern found", policy.ID)
		} else {
			resultLevel = sarifLevel
			resultText = fmt.Sprintf("API assurance failed for policy %s: Pattern not found", policy.ID)
		}
	} else {
		// For schema-based API checks
		if len(issues) == 0 {
			resultLevel = SARIFNote
			resultText = fmt.Sprintf("API validation passed for policy %s", policy.ID)
		} else {
			resultLevel = sarifLevel
			resultText = fmt.Sprintf("API validation failed for policy %s", policy.ID)
		}
	}

	timestamp := time.Now().Format(time.RFC3339)

	result := Result{
		RuleID: policy.ID,
		Level:  resultLevel,
		Message: Message{
			Text: resultText,
		},
		Locations: []Location{
			{
				PhysicalLocation: PhysicalLocation{
					ArtifactLocation: ArtifactLocation{
						URI: endpoint,
					},
				},
			},
		},
		Properties: map[string]string{
			"result-type":      "summary",
			"observe-run-id":   policy.RunID,
			"result-timestamp": timestamp,
			"name":             policy.Metadata.Name,
			"description":      policy.Metadata.Description,
			"msg-error":        policy.Metadata.MsgError,
			"msg-solution":     policy.Metadata.MsgSolution,
		},
	}

	sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, result)

	// Add detailed issues if any
	for _, issue := range issues {
		issueResult := Result{
			RuleID: policy.ID,
			Level:  sarifLevel,
			Message: Message{
				Text: fmt.Sprintf("API validation issue: %s", issue),
			},
			Locations: []Location{
				{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: ArtifactLocation{
							URI: endpoint,
						},
					},
				},
			},
			Properties: map[string]string{
				"result-type":      "detail",
				"observe-run-id":   policy.RunID,
				"result-timestamp": timestamp,
				"name":             policy.Metadata.Name,
				"description":      policy.Metadata.Description,
				"msg-error":        policy.Metadata.MsgError,
				"msg-solution":     policy.Metadata.MsgSolution,
			},
		}
		sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, issueResult)
	}

	return sarifReport, nil
}
