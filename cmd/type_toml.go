package cmd

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"github.com/BurntSushi/toml"
	xtoml "github.com/pelletier/go-toml"
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

	cuepolicy, _ := cueValue.Value().MarshalJSON()
	tomlcontent, _ := tomlCueValue.Value().MarshalJSON()

	err = cueValue.UnifyAccept(cueValue.Value(), tomlCueValue.Value()).Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating toml accept data against CUE schema: %v", err)
	}

	// keys present

	var jsonObj map[string]interface{}
	if err := json.Unmarshal([]byte(cuepolicy), &jsonObj); err != nil {
		return false, fmt.Sprintf("Error parsing policy: %v", err)
	}
	rootKeys := getJSONRootKeys(jsonObj)

	tree, err := xtoml.Load(string(tomlcontent))
	if err != nil {
		colorYellowBold.Println("│ Warning : TOML not valid")
		return true, ""
	} else {
		for _, key := range rootKeys {
			if isTOMLKeyAbsent(tree, key) {
				return false, fmt.Sprintf("TOML Key '%s' is absent", key)
			}
		}
	}

	return true, ""
}
