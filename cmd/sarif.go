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

func sarifLevelToInt(level SARIFLevel) int {
	switch level {
	case SARIFNone:
		return 0
	case SARIFNote:
		return 1
	case SARIFWarning:
		return 2
	case SARIFError:
		return 3
	default:
		return 99
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
	Tool Tool `json:"tool"`

	Results     []Result     `json:"results"`
	Invocations []Invocation `json:"invocations"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name            string      `json:"name"`
	Version         string      `json:"version"`
	FullName        string      `json:"fullName"`
	SemanticVersion string      `json:"semanticVersion"`
	InformationURI  string      `json:"informationUri"`
	Rules           []SARIFRule `json:"rules"`
}

type Result struct {
	RuleID     string           `json:"ruleId"`
	Level      SARIFLevel       `json:"level"`
	Message    Message          `json:"message"`
	Locations  []Location       `json:"locations,omitempty"`
	Properties ResultProperties `json:"properties,omitempty"`
}

type Message struct {
	Text string `json:"text"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           Region           `json:"region,omitempty"`
}

type ArtifactLocation struct {
	URI string `json:"uri,omitempty"`
}

type Region struct {
	StartLine   int     `json:"startLine,omitempty"`
	StartColumn int     `json:"startColumn,omitempty"`
	EndColumn   int     `json:"endColumn,omitempty"`
	Snippet     Snippet `json:"snippet,omitempty"`
}

type Snippet struct {
	Text string `json:"text,omitempty"`
}

type SARIFRule struct {
	ID                   string                `json:"id"`
	ShortDescription     ShortDescription      `json:"shortDescription"`
	FullDescription      *FullDescription      `json:"fullDescription,omitempty"`
	HelpURI              string                `json:"helpUri,omitempty"`
	Help                 *Help                 `json:"help,omitempty"`
	Properties           Properties            `json:"properties,omitempty"`
	DefaultLevel         string                `json:"defaultLevel,omitempty"`
	DefaultConfiguration *DefaultConfiguration `json:"defaultConfiguration,omitempty"`
}

type ShortDescription struct {
	Text string `json:"text"`
}

type FullDescription struct {
	Text string `json:"text"`
}

type Help struct {
	Text     string `json:"text"`
	Markdown string `json:"markdown,omitempty"`
}

type Properties struct {
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type DefaultConfiguration struct {
	Level string `json:"level,omitempty"`
}

type ResultProperties struct {
	ResourceType    string `json:"resource-type"`
	Property        string `json:"property"`
	ResultType      string `json:"result-type"`
	ObserveRunId    string `json:"observe-run-id"`
	ResultTimestamp string `json:"result-timestamp"`
	Environment     string `json:"environment"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	MsgError        string `json:"msg-error"`
	MsgSolution     string `json:"msg-solution"`
	SarifInt        int    `json:"sarif-int"`
}

type InvocationProperties struct {
	RunId             string `json:"run-id"`
	StartTime         string `json:"start-time"`
	EndTime           string `json:"end-time"`
	ExecutionTimeInMs string `json:"execution-time-ms"`
	Environment       string `json:"environment"`
	Debug             string `json:"debug"`
	ReportTimestamp   string `json:"report-timestamp"`
	HostData          string `json:"host-data"`
	HostFingerprint   string `json:"host-fingerprint"`
	ReportStatus      string `json:"report-status"`
	ReportCompliant   bool   `json:"report-compliant"`
}

type Invocation struct {
	ExecutionSuccessful bool                 `json:"executionSuccessful"`
	CommandLine         string               `json:"commandLine,omitempty"`
	Properties          InvocationProperties `json:"properties,omitempty"`
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
						FullName:        fmt.Sprintf("%s %s", "INTERCEPT", buildVersion),
						Name:            "INTERCEPT",
						Version:         smVersion,
						SemanticVersion: smVersion,
						InformationURI:  "https://intercept.cc",
						Rules:           policyData.SARIFRules,
					},
				},

				Results: []Result{},
				Invocations: []Invocation{
					{
						ExecutionSuccessful: true,
						Properties:          InvocationProperties{},
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
						Region: Region{
							StartLine:   1,
							StartColumn: 1,
							EndColumn:   1,
							Snippet: Snippet{
								Text: "N/A",
							},
						},
					},
				},
			},
			Properties: ResultProperties{
				ResultType:      "detail",
				ObserveRunId:    policy.RunID,
				ResultTimestamp: timestamp,
				Environment:     environment,
				Name:            policy.Metadata.Name,
				Description:     policy.Metadata.Description,
				MsgError:        policy.Metadata.MsgError,
				MsgSolution:     policy.Metadata.MsgSolution,
				SarifInt:        sarifLevelToInt(SARIFNote),
			},
		})
	} else {
		// Process ripgrep output and add results to SARIF report
		for _, rgOutput := range rgOutputs {
			if rgOutput.Type == "match" {

				sarifLevel := calculateSARIFLevel(policy, environment)

				for _, submatch := range rgOutput.Data.Submatches {
					matchText := submatch.Match.Text
					startColumn := strings.Index(rgOutput.Data.Lines.Text, matchText) + 1
					endColumn := startColumn + len(matchText)

					var message string

					if len(matchText) != 0 {
						message = fmt.Sprintf("Policy violation: %s Matched text: %s", policy.Metadata.Name, matchText)
					} else {
						message = fmt.Sprintf("Policy violation: %s", policy.Metadata.Name)
					}

					result := Result{
						RuleID: policy.ID,
						Level:  sarifLevel,
						Message: Message{
							Text: message,
						},
						Locations: []Location{
							{
								PhysicalLocation: PhysicalLocation{
									ArtifactLocation: ArtifactLocation{URI: rgOutput.Data.Path.Text},
									Region: Region{
										StartLine:   max(rgOutput.Data.LineNumber, 1),
										StartColumn: max(startColumn, 1),
										EndColumn:   max(endColumn, 1),
										Snippet: Snippet{
											Text: matchText,
										},
									},
								},
							},
						},
						Properties: ResultProperties{
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

					results = append(results, result)
				}
			}
		}
	}

	sarifReport.Runs[0].Results = results

	sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant = ComplianceStatus(sarifReport)

	if outputTypeMatrixConfig.LOG {
		PostResultsToComplianceLog(sarifReport)
	}

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
						FullName:        fmt.Sprintf("%s %s", "INTERCEPT", buildVersion),
						Name:            "INTERCEPT",
						Version:         smVersion,
						SemanticVersion: smVersion,
						InformationURI:  "https://intercept.cc",
						Rules:           policyData.SARIFRules,
					},
				},

				Results: []Result{},
				Invocations: []Invocation{
					{
						ExecutionSuccessful: true,
						Properties:          InvocationProperties{},
					},
				},
			},
		},
	}

	// Translate policy enforcement to SARIF level

	sarifLevel := calculateSARIFLevel(policy, environment)
	timestamp := time.Now().Format(time.RFC3339)

	result := Result{}
	if status == "FOUND" {
		// For assure policies, we report based on the status
		result = Result{
			RuleID: policy.ID,
			Level:  SARIFNote,
			Message: Message{
				Text: fmt.Sprintf("Assure policy %s: Pattern %s", policy.Metadata.Name, status),
			},
			Properties: ResultProperties{
				ResultType:      "detail",
				ObserveRunId:    policy.RunID,
				ResultTimestamp: timestamp,
				Environment:     environment,
				Name:            policy.Metadata.Name,
				Description:     policy.Metadata.Description,
				MsgError:        policy.Metadata.MsgError,
				MsgSolution:     policy.Metadata.MsgSolution,
				SarifInt:        sarifLevelToInt(SARIFNote),
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
			Properties: ResultProperties{
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
	}

	if len(rgOutputs) > 0 && status == "FOUND" {
		// If matches were found, include location information
		for _, rgOutput := range rgOutputs {
			if rgOutput.Type == "match" {
				result.Locations = append(result.Locations, Location{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: ArtifactLocation{URI: rgOutput.Data.Path.Text},
						Region: Region{
							StartLine:   max(rgOutput.Data.LineNumber, 1),
							StartColumn: 1,
							EndColumn:   max(len(rgOutput.Data.Lines.Text), 1),
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
						Snippet: Snippet{
							Text: "N/A",
						},
					},
				},
			},
		}
	}

	sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, result)

	sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant = ComplianceStatus(sarifReport)

	if outputTypeMatrixConfig.LOG {
		PostResultsToComplianceLog(sarifReport)
	}

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
						FullName:        fmt.Sprintf("%s %s", "INTERCEPT", buildVersion),
						Name:            "INTERCEPT",
						Version:         smVersion,
						SemanticVersion: smVersion,
						InformationURI:  "https://intercept.cc",
						Rules:           policyData.SARIFRules,
					},
				},

				Results:     []Result{},
				Invocations: []Invocation{{ExecutionSuccessful: true, Properties: InvocationProperties{}}},
			},
		},
	}

	sarifLevel := calculateSARIFLevel(policy, environment)

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
							Region: Region{
								StartLine:   1,
								StartColumn: 1,
								EndColumn:   1,
								Snippet: Snippet{
									Text: "N/A",
								},
							},
						},
					},
				},
				Properties: ResultProperties{
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
					Region: Region{
						StartLine:   1,
						StartColumn: 1,
						EndColumn:   1,
						Snippet: Snippet{
							Text: "N/A",
						},
					},
				},
			},
		},
		Properties: ResultProperties{
			ResultType:      "detail",
			ObserveRunId:    policy.RunID,
			ResultTimestamp: timestamp,
			Environment:     environment,
			Name:            policy.Metadata.Name,
			Description:     policy.Metadata.Description,
			MsgError:        policy.Metadata.MsgError,
			MsgSolution:     policy.Metadata.MsgSolution,
			SarifInt:        sarifLevelToInt(detailSarif),
		},
	}
	sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, summaryResult)

	sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant = ComplianceStatus(sarifReport)

	if outputTypeMatrixConfig.LOG {
		PostResultsToComplianceLog(sarifReport)
	}

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
						FullName:        fmt.Sprintf("%s %s", "INTERCEPT", buildVersion),
						Name:            "INTERCEPT",
						Version:         smVersion,
						SemanticVersion: smVersion,
						InformationURI:  "https://intercept.cc",
						Rules:           policyData.SARIFRules,
					},
				},

				Results: []Result{},
				Invocations: []Invocation{
					{
						ExecutionSuccessful: true,
						CommandLine:         commandLine,
						Properties: InvocationProperties{
							RunId:             intercept_run_id,
							StartTime:         perf.StartTime.Format(time.RFC3339),
							EndTime:           perf.EndTime.Format(time.RFC3339),
							ExecutionTimeInMs: fmt.Sprintf("%d", perf.Delta.Milliseconds()),
							Environment:       environment,
							Debug:             fmt.Sprintf("%v", debugOutput),
							ReportTimestamp:   timestamp,
							HostData:          hostData,
							HostFingerprint:   hostFingerprint,
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
		mergedReport.Runs[0].Invocations[0].Properties.ReportStatus = "compliant"
		mergedReport.Runs[0].Invocations[0].Properties.ReportCompliant = true
	} else {
		mergedReport.Runs[0].Invocations[0].Properties.ReportStatus = "non-compliant"
		mergedReport.Runs[0].Invocations[0].Properties.ReportCompliant = false
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

	if outputTypeMatrixConfig.LOG {
		PostReportToComplianceLog(mergedReport)
	}

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
	sarifReport := SARIFReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						FullName:        fmt.Sprintf("%s %s", "INTERCEPT", buildVersion),
						Name:            "INTERCEPT",
						Version:         smVersion,
						SemanticVersion: smVersion,
						InformationURI:  "https://intercept.cc",
						Rules:           policyData.SARIFRules,
					},
				},

				Results:     results,
				Invocations: []Invocation{{ExecutionSuccessful: true, Properties: InvocationProperties{}}},
			},
		},
	}

	sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant = ComplianceStatus(sarifReport)

	if outputTypeMatrixConfig.LOG {
		PostResultsToComplianceLog(sarifReport)
	}

	return sarifReport

}
func GenerateAPISARIFReport(policy Policy, endpoint string, matchFound bool, issues []string) (SARIFReport, error) {
	sarifReport := SARIFReport{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						FullName:        fmt.Sprintf("%s %s", "INTERCEPT", buildVersion),
						Name:            "INTERCEPT",
						Version:         smVersion,
						SemanticVersion: smVersion,
						InformationURI:  "https://intercept.cc",
						Rules:           policyData.SARIFRules,
					},
				},

				Results:     []Result{},
				Invocations: []Invocation{{ExecutionSuccessful: true, Properties: InvocationProperties{}}},
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
		Properties: ResultProperties{
			ResultType:      "summary",
			ObserveRunId:    policy.RunID,
			ResultTimestamp: timestamp,
			Environment:     environment,
			Name:            policy.Metadata.Name,
			Description:     policy.Metadata.Description,
			MsgError:        policy.Metadata.MsgError,
			MsgSolution:     policy.Metadata.MsgSolution,
			SarifInt:        sarifLevelToInt(resultLevel),
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
						Region: Region{
							StartLine:   1,
							StartColumn: 1,
							EndColumn:   1,
							Snippet: Snippet{
								Text: "N/A",
							},
						},
					},
				},
			},
			Properties: ResultProperties{
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
		sarifReport.Runs[0].Results = append(sarifReport.Runs[0].Results, issueResult)
	}

	sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant = ComplianceStatus(sarifReport)

	if outputTypeMatrixConfig.LOG {
		PostResultsToComplianceLog(sarifReport)
	}

	return sarifReport, nil
}

func ComplianceStatus(sarifReport SARIFReport) bool {

	for _, result := range sarifReport.Runs[0].Results {
		if result.Level == "error" {
			return false
		}
		if result.Level == "warning" {
			return false
		}
	}
	return true
}
