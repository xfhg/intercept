package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/open-policy-agent/opa/topdown"
)

func ProcessRegoType(policy Policy, targetDir string, filePaths []string) error {
	policyContent, err := os.ReadFile(policy.Rego.PolicyFile)
	if err != nil {
		log.Error().Err(err).Str("file", policy.Rego.PolicyFile).Msg("Error reading policy file")
		return fmt.Errorf("error reading policy file %s: %w", policy.Rego.PolicyFile, err)
	}
	packageName, err := extractPackageName(string(policyContent))
	if err != nil {
		log.Error().Err(err).Str("policy", policy.ID).Msg("Error extracting package name")
		return fmt.Errorf("error extracting package name for policy %s: %w", policy.ID, err)
	}

	queryPackage := extractQueryPackage(policy.Rego.PolicyQuery)
	if queryPackage != packageName {
		log.Error().
			Str("policy", policy.ID).
			Str("queryPackage", queryPackage).
			Str("packageName", packageName).
			Msg("Policy query package does not match Rego file package")
		return fmt.Errorf("policy query package (%s) does not match Rego file package (%s) for policy %s", queryPackage, packageName, policy.ID)
	}

	policyData, err := readJSONFile(policy.Rego.PolicyData)
	if err != nil {
		log.Error().Err(err).Str("file", policy.Rego.PolicyData).Msg("Error reading policy data file")
		return fmt.Errorf("error reading policy data file %s: %w", policy.Rego.PolicyData, err)
	}

	var allResults []Result

	for _, filePath := range filePaths {
		log.Debug().Str("policy", policy.ID).Str("file", filePath).Msg("Processing REGO policy")

		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Error().Err(err).Str("file", filePath).Msg("Error reading input file")
			return fmt.Errorf("error reading input file %s: %w", filePath, err)
		}

		var input map[string]interface{}
		if json.Valid(fileContent) {
			// If the file is valid JSON, use it directly as input
			if err := json.Unmarshal(fileContent, &input); err != nil {
				log.Error().Err(err).Str("file", filePath).Msg("Error parsing input JSON")
				return fmt.Errorf("error parsing input JSON %s: %w", filePath, err)
			}
		} else {
			// For non-JSON files (like nginx config), use the existing input structure
			input = map[string]interface{}{
				"content": string(fileContent),
				"lines":   strings.Split(string(fileContent), "\n"),
				"blocks":  parseBlocks(string(fileContent)),
				"path":    filePath,
			}
		}

		tracer := topdown.NewBufferTracer()

		r := rego.New(
			rego.Query(fmt.Sprintf("data.%s", packageName)),
			rego.Module("policy.rego", string(policyContent)),
			rego.Input(input),
			rego.Store(inmem.NewFromObject(policyData)),
			rego.Tracer(tracer),
		)

		ctx := context.Background()

		query, err := r.PrepareForEval(ctx)
		if err != nil {
			log.Error().Err(err).Str("policy", policy.ID).Msg("Error preparing query")
			return fmt.Errorf("error preparing query for policy %s: %w", policy.ID, err)
		}

		results, err := query.Eval(ctx, rego.EvalTracer(tracer))
		if err != nil {
			log.Error().Err(err).Str("policy", policy.ID).Msg("Error evaluating policy")
			return fmt.Errorf("error evaluating policy %s: %w", policy.ID, err)
		}

		compliant, violations, err := checkCompliance(results)
		if err != nil {
			log.Error().Err(err).Str("policy", policy.ID).Msg("Error checking compliance")
			return fmt.Errorf("error checking compliance for policy %s: %w", policy.ID, err)
		}

		// Ensure SARIF level is "Note" if compliant is true
		sarifLevel := SARIFNote
		if !compliant {
			sarifLevel = calculateSARIFLevel(policy, environment)
		}

		timestamp := time.Now().Format(time.RFC3339)

		result := Result{
			RuleID: policy.ID,
			Level:  sarifLevel,
			Message: Message{
				Text: fmt.Sprintf("Policy %s %s for file %s with %d violations", policy.ID, map[bool]string{true: "passed", false: "failed"}[compliant], filePath, len(violations)),
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
			Properties: ResultProperties{
				ResultType:      "summary",
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

		allResults = append(allResults, result)

		for _, violation := range violations {
			result_detail := Result{
				RuleID: policy.ID,
				Level:  sarifLevel,
				Message: Message{
					Text: fmt.Sprintf("%s for file %s : Violation [ %s ] ", policy.ID, filePath, violation),
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

			allResults = append(allResults, result_detail)
		}

		// log.Debug().
		// 	Str("policy", policy.ID).
		// 	Str("file", filePath).
		// 	Bool("compliant", compliant).
		// 	Msg("Policy evaluation result")

		// // Log detailed results if needed
		// for i, res := range results {
		// 	for j, expr := range res.Expressions {
		// 		log.Debug().
		// 			Str("policy", policy.ID).
		// 			Str("file", filePath).
		// 			Int("result", i).
		// 			Int("expression", j).
		// 			Str("text", expr.Text).
		// 			Interface("value", expr.Value).
		// 			Msg("Evaluation expression result")
		// 	}
		// }
	}

	// Create a single SARIF report for all files
	sarifReport := createSARIFReport(allResults)

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

	log.Debug().Msgf("Policy %s processed. SARIF report written to: %s ", policy.ID, sarifOutputFile)

	return nil
}

func readJSONFile(filePath string) (map[string]interface{}, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
func checkCompliance(results rego.ResultSet) (bool, []string, error) {
	if len(results) == 0 {
		return false, nil, fmt.Errorf("no results returned from policy evaluation")
	}

	for _, result := range results {
		for _, expression := range result.Expressions {
			if value, ok := expression.Value.(map[string]interface{}); ok {
				allow, allowOk := value["allow"].(bool)
				violations, violationsOk := value["violations"].([]interface{})

				if allowOk {
					if violationsOk {
						violationMsgs := make([]string, len(violations))
						for i, v := range violations {
							violationMsgs[i] = fmt.Sprint(v)
						}
						return allow, violationMsgs, nil
					}
					return allow, nil, nil
				}
			}
		}
	}

	return false, nil, fmt.Errorf("unexpected result format from policy evaluation")
}

func parseBlocks(content string) []map[string]interface{} {
	var blocks []map[string]interface{}
	lines := strings.Split(content, "\n")
	var currentBlock map[string]interface{}
	var blockContent []string
	var inBlock bool

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasSuffix(trimmedLine, "{") {
			if inBlock {
				// Nested block, add to current block's content
				blockContent = append(blockContent, line)
			} else {
				// New top-level block
				inBlock = true
				currentBlock = make(map[string]interface{})
				currentBlock["directive"] = strings.TrimSuffix(trimmedLine, " {")
				blockContent = []string{}
			}
		} else if trimmedLine == "}" {
			if len(blockContent) > 0 {
				// End of a nested block
				blockContent = append(blockContent, line)
			} else {
				// End of top-level block
				inBlock = false
				currentBlock["block"] = blockContent
				blocks = append(blocks, currentBlock)
				currentBlock = nil
			}
		} else if inBlock {
			blockContent = append(blockContent, line)
		}
	}

	return blocks
}

func extractPackageName(policyContent string) (string, error) {
	re := regexp.MustCompile(`(?m)^package\s+(\w+)`)
	matches := re.FindStringSubmatch(policyContent)
	if len(matches) < 2 {
		return "", fmt.Errorf("package name not found in policy content")
	}
	return matches[1], nil
}

func extractQueryPackage(query string) string {
	parts := strings.Split(query, ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}
