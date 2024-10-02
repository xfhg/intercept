package cmd

import (
	"fmt"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
)

func validateContentAndCUE(content []byte, cueContent string, contentType string, strictSchema bool, policyID string) (bool, []string) {
	var issues []string

	// Convert content to JSON (implementation depends on contentType)
	jsonContent, err := convertToJSON(content, contentType)
	if err != nil {
		issues = append(issues, fmt.Sprintf("Error converting %s to JSON: %v", contentType, err))
		return false, issues
	}

	ctx := cuecontext.New()
	cueValue := ctx.CompileString(cueContent)
	if cueValue.Err() != nil {
		issues = append(issues, fmt.Sprintf("Error compiling CUE content: %v", cueValue.Err()))
		return false, issues
	}

	jsonCueValue := ctx.CompileBytes(jsonContent)
	if jsonCueValue.Err() != nil {
		issues = append(issues, fmt.Sprintf("Error compiling JSON data to CUE value: %v", jsonCueValue.Err()))
		return false, issues
	}

	unified := cueValue.Unify(jsonCueValue)
	if err := unified.Validate(); err != nil {
		errors := extractCUEErrors(err)
		issues = append(issues, errors...)
	}

	missingFields, extraFields := validateSchema(cueValue, jsonCueValue)

	// missingFields := findMissingFields(cueValue, jsonCueValue)
	for _, field := range missingFields {
		issues = append(issues, fmt.Sprintf("Missing required field: %s", field))
	}

	if strictSchema {
		// extraFields := findExtraFields(cueValue, jsonCueValue)
		for _, field := range extraFields {
			issues = append(issues, fmt.Sprintf("Extra field not defined in schema: %s", field))
		}
	}

	log.Warn().Str("policy", policyID).Msgf("Missing fields: %s", missingFields)
	log.Warn().Str("policy", policyID).Msgf("Extra fields: %s", extraFields)

	return len(issues) == 0, issues
}

func extractCUEErrors(err error) []string {
	var errs []string
	for _, e := range errors.Errors(err) {
		errs = append(errs, fmt.Sprintf("Validation error at %v: %v", e.Path(), e.Error()))
	}
	return errs
}

// This function is shared across all policy types
func generateSchemaResults(policy Policy, filePath string, valid bool, issues []string, patched bool) []Result {
	var results []Result

	sarifLevel := calculateSARIFLevel(policy, environment)

	if !valid {
		for _, issue := range issues {
			result := Result{
				RuleID: policy.ID,
				Level:  sarifLevel,
				Message: Message{
					Text: fmt.Sprintf("Schema validation issue: %s", issue),
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
			}
			results = append(results, result)
		}
	}

	// Add a summary result
	summaryText := fmt.Sprintf("Schema validation %s for policy %s",
		map[bool]string{true: "passed", false: "failed"}[valid],
		policy.ID)
	if patched {
		summaryText += " (content patched)"
	}

	summaryResult := Result{
		RuleID: policy.ID,
		Level:  map[bool]SARIFLevel{true: SARIFNote, false: sarifLevel}[valid],
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
	}
	results = append(results, summaryResult)

	return results
}
