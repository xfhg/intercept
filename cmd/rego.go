package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	policyData, err := readJSONFile(policy.Rego.PolicyData)
	if err != nil {
		log.Error().Err(err).Str("file", policy.Rego.PolicyData).Msg("Error reading policy data file")
		return fmt.Errorf("error reading policy data file %s: %w", policy.Rego.PolicyData, err)
	}

	var allResults []Result

	for _, filePath := range filePaths {
		log.Debug().Str("policy", policy.ID).Str("file", filePath).Msg("Processing REGO policy")

		input, err := readJSONFile(filePath)
		if err != nil {
			log.Error().Err(err).Str("file", filePath).Msg("Error reading input file")
			return fmt.Errorf("error reading input file %s: %w", filePath, err)
		}

		tracer := topdown.NewBufferTracer()

		r := rego.New(
			rego.Query(policy.Rego.PolicyQuery),
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

		compliant, err := checkCompliance(results)
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
				Text: fmt.Sprintf("Policy %s %s for file %s", policy.ID, map[bool]string{true: "passed", false: "failed"}[compliant], filePath),
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

		log.Debug().
			Str("policy", policy.ID).
			Str("file", filePath).
			Bool("compliant", compliant).
			Msg("Policy evaluation result")

		// Log detailed results if needed
		for i, res := range results {
			for j, expr := range res.Expressions {
				log.Debug().
					Str("policy", policy.ID).
					Str("file", filePath).
					Int("result", i).
					Int("expression", j).
					Str("text", expr.Text).
					Interface("value", expr.Value).
					Msg("Evaluation expression result")
			}
		}
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

func checkCompliance(results rego.ResultSet) (bool, error) {
	if len(results) == 0 {
		return false, fmt.Errorf("no results returned from policy evaluation")
	}

	for i, result := range results {
		log.Debug().Msgf("Result %d: %v ", i, result)
	}

	for _, result := range results {
		for _, expression := range result.Expressions {
			switch value := expression.Value.(type) {
			case bool:
				return value, nil
			case []interface{}:
				// If the result is an array, check if it's non-empty
				return len(value) > 0, nil
			case map[string]interface{}:
				// If the result is an object, check if it's non-empty
				return len(value) > 0, nil
			default:
				// For any other type, consider it a pass if it's not nil
				return value != nil, nil
			}
		}
	}

	return false, fmt.Errorf("unexpected result format from policy evaluation")
}
