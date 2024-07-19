package cmd

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"gopkg.in/ini.v1"
)

func validateINIAndCUEContent(iniContent string, cueContent string) (bool, string) {
	// Parse INI content
	cfg, err := ini.Load([]byte(iniContent))
	if err != nil {
		return false, fmt.Sprintf("error unmarshaling INI data 3: %v", err)
	}

	// Convert INI data to a map for JSON marshaling
	iniMap := make(map[string]interface{})
	for _, section := range cfg.Sections() {
		sectionMap := make(map[string]interface{})
		for _, key := range section.Keys() {
			sectionMap[key.Name()] = key.Value()
		}
		iniMap[section.Name()] = sectionMap
	}

	// Marshal INI data to JSON
	jsonData, err := json.Marshal(iniMap)
	if err != nil {
		return false, fmt.Sprintf("error marshaling INI object to JSON: %v", err)
	}

	// Compile and validate CUE content
	cueCtx := cuecontext.New()
	binst := cueCtx.CompileBytes([]byte(cueContent))
	if err != nil {
		return false, fmt.Sprintf("error compiling CUE content: %v", err)
	}

	cueValue := binst
	err = cueValue.Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating CUE value: %v", err)
	}

	// Compile JSON data to CUE value
	iniCueValue := cueCtx.CompileBytes(jsonData)
	if err != nil {
		return false, fmt.Sprintf("error compiling JSON data to CUE value: %v", err)
	}

	// Validate INI data against CUE schema
	err = cueValue.Unify(iniCueValue).Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating INI data against CUE schema: %v", err)
	}

	return true, ""
}
