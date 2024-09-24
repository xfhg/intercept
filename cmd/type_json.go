package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

func ProcessJSONType(policy Policy, targetDir string, filePaths []string) error {
	var allResults []Result

	for _, filePath := range filePaths {
		jsonContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Error().Err(err).Msg("error reading JSON file %s")
			return fmt.Errorf("error reading JSON file %s: %w", filePath, err)
		}

		cueContent := policy.Schema.Structure
		valid, issues := validateContentAndCUE(jsonContent, cueContent, "json", policy.Schema.Strict, policy.ID)

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

	if lLog {
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

func ProcessJSONTypeWithPatch(policy Policy, targetDir string, filePaths []string) error {
	var allResults []Result

	// Create _patched directory if it doesn't exist
	patchedDir := "_patched"
	if err := os.MkdirAll(patchedDir, 0755); err != nil {
		log.Error().Err(err).Msg("error creating _patched directory")
		return fmt.Errorf("error creating _patched directory: %w", err)
	}

	for _, filePath := range filePaths {
		jsonContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Error().Err(err).Msg("error reading JSON file %s")
			return fmt.Errorf("error reading JSON file %s: %w", filePath, err)
		}

		cueContent := policy.Schema.Structure
		valid, issues, patchedContent := validateAndPatchContentWithCUE(jsonContent, cueContent)

		// Generate results for this file
		fileResults := generateSchemaResults(policy, filePath, valid, issues, patchedContent != nil)
		allResults = append(allResults, fileResults...)

		if !valid {
			log.Debug().Msgf("Policy %s validation failed for file %s: ", policy.ID, filePath)
			for _, issue := range issues {
				log.Debug().Msgf("- %s ", issue)
			}

			if patchedContent != nil {
				// Create patched file in _patched directory
				patchedFileName := fmt.Sprintf("%s.patched.%s", filepath.Base(filePath), "json")
				patchedFilePath := filepath.Join(patchedDir, patchedFileName)
				err = os.WriteFile(patchedFilePath, patchedContent, 0644)
				if err != nil {
					log.Error().Err(err).Msg("error writing patched file %s")
					return fmt.Errorf("error writing patched file %s: %w", patchedFilePath, err)
				}
				log.Debug().Msgf("Patched content written to: %s ", patchedFilePath)
			}
		} else {
			log.Debug().Msgf("Policy %s validation passed for file %s ", policy.ID, filePath)
		}
	}

	// Create a single SARIF report for all files
	sarifReport := createSARIFReport(allResults)

	// Write SARIF report
	err := writeSARIFReport(policy.ID, sarifReport)
	if err != nil {
		log.Error().Err(err).Msg("error writing SARIF report for policy %s")
		return fmt.Errorf("error writing SARIF report for policy %s: %w", policy.ID, err)
	}

	return nil
}
