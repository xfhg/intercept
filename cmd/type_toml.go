package cmd

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/BurntSushi/toml"
)

func validateTOMLAndCUEContent(tomlContent string, cueContent string) (bool, string) {
	var tomlObj interface{}
	err := toml.Unmarshal([]byte(tomlContent), &tomlObj)
	if err != nil {
		return false, fmt.Sprintf("error unmarshaling toml data: %v", err)
	}

	var r cue.Runtime
	binst, err := r.Compile("", cueContent)
	if err != nil {
		return false, fmt.Sprintf("error compiling CUE content: %v", err)
	}

	cueValue := binst.Value()
	err = cueValue.Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating CUE value: %v", err)
	}

	jsonData, err := json.Marshal(tomlObj)
	if err != nil {
		return false, fmt.Sprintf("error marshaling toml object to JSON: %v", err)
	}

	tomlCueValue, err := r.Compile("", string(jsonData))
	if err != nil {
		return false, fmt.Sprintf("error compiling JSON data to CUE value: %v", err)
	}

	err = cueValue.Unify(tomlCueValue.Value()).Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating toml data against CUE schema: %v", err)
	}

	return true, ""
}
