package cmd

import (
	"fmt"
	"os"
)

func ProcessTOMLType(policy Policy, targetDir string, filePaths []string) error {
	var allResults []Result

	for _, filePath := range filePaths {
		tomlContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Error().Err(err).Msg("error reading TOML file %s")
			return fmt.Errorf("error reading TOML file %s: %w", filePath, err)
		}

		cueContent := policy.Schema.Structure
		valid, issues := validateContentAndCUE(tomlContent, cueContent, "toml", policy.Schema.Strict)

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
