package cmd

import (
	"fmt"
	"os"
)

func ProcessYAMLType(policy Policy, targetDir string, filePaths []string) error {
	var allResults []Result

	for _, filePath := range filePaths {
		yamlContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Error().Err(err).Msg("error reading YAML file %s")
			return fmt.Errorf("error reading YAML file %s: %w", filePath, err)
		}

		cueContent := policy.Schema.Structure
		valid, issues := validateContentAndCUE(yamlContent, cueContent, "yaml", policy.Schema.Strict, policy.ID)

		// Generate results for this file
		fileResults := generateSchemaResults(policy, filePath, valid, issues, false)
		allResults = append(allResults, fileResults...)

		if !valid {
			log.Debug().Msgf("Policy %s validation failed for file %s: ", policy.ID, filePath)
			for _, issue := range issues {
				log.Debug().Msgf("- %s ", issue)
			}
		} else {
			log.Debug().Msgf("Policy %s validation passed for file %s ", policy.ID, filePath)
		}
	}

	// Create a single SARIF report for all files
	sarifReport := createSARIFReport(allResults)

	if outputTypeMatrixConfig.LOG {
		PostResultsToComplianceLog(sarifReport)
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

	log.Debug().Msgf("Policy %s processed. SARIF report written to: %s ", policy.ID, sarifOutputFile)

	return nil
}
