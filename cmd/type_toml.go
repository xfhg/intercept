package cmd

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"github.com/pelletier/go-toml/v2"
)

func validateTOMLAndCUEContent(tomlContent, cueContent string) (bool, string) {
	// Parse TOML content
	var tomlObj map[string]interface{}
	if err := toml.Unmarshal([]byte(tomlContent), &tomlObj); err != nil {
		return false, fmt.Sprintf("error unmarshaling TOML data: %v", err)
	}

	// Create a new CUE context
	ctx := cuecontext.New()

	// Compile CUE content
	cueValue := ctx.CompileString(cueContent)
	if cueValue.Err() != nil {
		return false, fmt.Sprintf("error compiling CUE content: %v", cueValue.Err())
	}

	if err := cueValue.Validate(cue.Concrete(true)); err != nil {
		return false, fmt.Sprintf("error validating CUE value: %v", err)
	}

	// Convert TOML to JSON
	jsonData, err := json.Marshal(tomlObj)
	if err != nil {
		return false, fmt.Sprintf("error marshaling TOML object to JSON: %v", err)
	}

	// Parse JSON data to CUE value
	tomlCueValue := ctx.CompileBytes(jsonData)
	if tomlCueValue.Err() != nil {
		return false, fmt.Sprintf("error compiling JSON data to CUE value: %v", tomlCueValue.Err())
	}

	// Validate TOML data against CUE schema
	unifiedValue := cueValue.Unify(tomlCueValue)
	if err := unifiedValue.Validate(cue.Concrete(true)); err != nil {
		return false, fmt.Sprintf("error validating TOML data against CUE schema: %v", err)
	}

	// Validate TOML accept data against CUE schema
	acceptedValue := cueValue.UnifyAccept(cueValue, tomlCueValue)
	if err := acceptedValue.Validate(cue.Concrete(true)); err != nil {
		return false, fmt.Sprintf("error validating TOML accept data against CUE schema: %v", err)
	}

	// Check for missing keys in TOML
	cuePolicyJSON, err := cueValue.MarshalJSON()
	if err != nil {
		return false, fmt.Sprintf("error marshaling CUE policy to JSON: %v", err)
	}

	var jsonObj map[string]interface{}
	if err := json.Unmarshal(cuePolicyJSON, &jsonObj); err != nil {
		return false, fmt.Sprintf("error parsing policy: %v", err)
	}

	rootKeys := getJSONRootKeys(jsonObj)

	// Check for missing keys in TOML
	for _, key := range rootKeys {
		if _, exists := tomlObj[key]; !exists {
			return false, fmt.Sprintf("TOML key '%s' is absent", key)
		}
	}

	return true, ""
}
