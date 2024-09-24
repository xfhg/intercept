package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

func ProcessINIType(policy Policy, targetDir string, filePaths []string) error {
	var allResults []Result

	for _, filePath := range filePaths {
		iniContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Error().Err(err).Msg("error reading INI file %s")
			return fmt.Errorf("error reading INI file %s: %w", filePath, err)
		}

		cueContent := policy.Schema.Structure
		valid, issues := validateContentAndCUE(iniContent, cueContent, "ini", policy.Schema.Strict, policy.ID)

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
func iniToJSONLike(cfg *ini.File) string {
	result := make(map[string]interface{})

	for _, section := range cfg.Sections() {
		if section.Name() == "DEFAULT" {
			// Flatten DEFAULT section
			for _, key := range section.Keys() {
				result[key.Name()] = key.Value()
			}
		} else {
			sectionMap := make(map[string]interface{})
			for _, key := range section.Keys() {
				sectionMap[key.Name()] = parseValue(key.Value())
			}
			result[section.Name()] = sectionMap
		}
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Debug().Msgf("Error marshalling JSON: %v ", err)
		return ""
	}
	return string(jsonBytes)
}

func parseValue(value string) interface{} {
	// Try to parse as bool
	if strings.ToLower(value) == "true" {
		return true
	}
	if strings.ToLower(value) == "false" {
		return false
	}

	// Try to parse as int
	if intValue, err := json.Number(value).Int64(); err == nil {
		return intValue
	}

	// Try to parse as float
	if floatValue, err := json.Number(value).Float64(); err == nil {
		return floatValue
	}

	// Return as string if all else fails
	return value
}
