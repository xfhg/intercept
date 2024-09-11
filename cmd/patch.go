package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
)

func validateAndPatchContentWithCUE(content []byte, cueContent string) (bool, []string, []byte) {
	var issues []string

	ctx := cuecontext.New()
	cueValue := ctx.CompileString(cueContent)
	if cueValue.Err() != nil {
		issues = append(issues, fmt.Sprintf("Error compiling CUE content: %v", cueValue.Err()))
		return false, issues, nil
	}

	var jsonData map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(content))
	d.UseNumber()
	if err := d.Decode(&jsonData); err != nil {
		issues = append(issues, fmt.Sprintf("Error parsing JSON content: %v", err))
		return false, issues, nil
	}

	jsonData = convertJSONNumbers(jsonData).(map[string]interface{})

	jsonCueValue := ctx.Encode(jsonData)
	if jsonCueValue.Err() != nil {
		issues = append(issues, fmt.Sprintf("Error encoding JSON data to CUE value: %v", jsonCueValue.Err()))
		return false, issues, nil
	}

	unified := cueValue.Unify(jsonCueValue)
	if err := unified.Validate(); err != nil {
		errors := extractCUEErrors(err)
		issues = append(issues, errors...)
	}

	patchedData := make(map[string]interface{})
	patchData(cueValue, jsonData, patchedData)

	if len(patchedData) > 0 {
		// Apply patches
		applyPatches(jsonData, patchedData)

		patchedContent, err := json.MarshalIndent(jsonData, "", "  ")
		if err != nil {
			issues = append(issues, fmt.Sprintf("Error marshaling patched content: %v", err))
			return false, issues, nil
		}

		issues = append(issues, "Content patched according to CUE schema")
		return false, issues, patchedContent
	}

	return len(issues) == 0, issues, nil
}

func convertJSONNumbers(v interface{}) interface{} {
	switch v := v.(type) {
	case map[string]interface{}:
		m := make(map[string]interface{})
		for k, val := range v {
			m[k] = convertJSONNumbers(val)
		}
		return m
	case []interface{}:
		for i, val := range v {
			v[i] = convertJSONNumbers(val)
		}
		return v
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}
		if f, err := v.Float64(); err == nil {
			return f
		}
		return v.String()
	default:
		return v
	}
}

func patchData(cueValue cue.Value, jsonData, patchedData map[string]interface{}) {
	iter, _ := cueValue.Fields()
	for iter.Next() {
		label := iter.Label()
		value := iter.Value()

		if value.Kind() == cue.StructKind {
			if nestedJSON, ok := jsonData[label].(map[string]interface{}); ok {
				nestedPatch := make(map[string]interface{})
				patchData(value, nestedJSON, nestedPatch)
				if len(nestedPatch) > 0 {
					patchedData[label] = nestedPatch
				}
			} else {
				nestedPatch := make(map[string]interface{})
				patchData(value, make(map[string]interface{}), nestedPatch)
				if len(nestedPatch) > 0 {
					patchedData[label] = nestedPatch
				}
			}
		} else {
			cueVal := extractCueValue(value)
			if cueVal != nil {
				jsonVal, exists := jsonData[label]
				if !exists || !compareValues(cueVal, jsonVal) {
					patchedData[label] = cueVal
				}
			}
		}
	}
}

func extractCueValue(cueValue cue.Value) interface{} {
	switch cueValue.Kind() {
	case cue.StringKind:
		str, _ := cueValue.String()
		return str
	case cue.IntKind:
		i, _ := cueValue.Int64()
		return i
	case cue.FloatKind:
		f, _ := cueValue.Float64()
		return f
	case cue.BoolKind:
		b, _ := cueValue.Bool()
		return b
	default:
		return nil
	}
}

func compareValues(cueVal, jsonVal interface{}) bool {
	switch cueTyped := cueVal.(type) {
	case string:
		jsonTyped, ok := jsonVal.(string)
		return ok && cueTyped == jsonTyped
	case int64:
		switch jsonTyped := jsonVal.(type) {
		case int64:
			return jsonTyped == cueTyped
		case float64:
			return int64(jsonTyped) == cueTyped
		}
	case float64:
		jsonTyped, ok := jsonVal.(float64)
		return ok && cueTyped == jsonTyped
	case bool:
		jsonTyped, ok := jsonVal.(bool)
		return ok && cueTyped == jsonTyped
	}
	return false
}

func applyPatches(jsonData, patches map[string]interface{}) {
	for key, patchValue := range patches {
		if nestedPatch, ok := patchValue.(map[string]interface{}); ok {
			if nestedJSON, exists := jsonData[key].(map[string]interface{}); exists {
				applyPatches(nestedJSON, nestedPatch)
			} else {
				jsonData[key] = patchValue
			}
		} else {
			jsonData[key] = patchValue
		}
	}
}
func processGenericType(policy Policy, filePaths []string, fileType string) error {
	var allResults []Result

	// Create _patched directory if it doesn't exist
	patchedDir := "_patched"
	if err := os.MkdirAll(patchedDir, 0755); err != nil {
		log.Error().Err(err).Msg("error creating _patched directory")
		return fmt.Errorf("error creating _patched directory: %w", err)
	}

	for _, filePath := range filePaths {
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Error().Err(err).Msg("error reading %s file %s")
			return fmt.Errorf("error reading %s file %s: %w", fileType, filePath, err)
		}

		// Convert content to JSON
		jsonContent, err := convertToJSON(content, fileType)
		if err != nil {
			log.Error().Err(err).Msg("error converting %s to JSON for file %s")
			return fmt.Errorf("error converting %s to JSON for file %s: %w", fileType, filePath, err)
		}

		cueContent := policy.Schema.Structure
		valid, issues, patchedJSONContent := validateAndPatchContentWithCUE(jsonContent, cueContent)

		var patchedContent []byte
		if patchedJSONContent != nil {
			// Convert patched JSON back to original format
			patchedContent, err = convertFromJSON(patchedJSONContent, fileType)
			if err != nil {
				log.Error().Err(err).Msg("error converting patched JSON back to %s for file %s")
				return fmt.Errorf("error converting patched JSON back to %s for file %s: %w", fileType, filePath, err)
			}
		}

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
				patchedFileName := fmt.Sprintf("%s.patched.%s", filepath.Base(filePath), fileType)
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
