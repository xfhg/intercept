package cmd

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"gopkg.in/yaml.v3"
)

// func validateYAMLAgainstCUE(yamlFile string, cueFile string) (bool, string) {
// 	yamlData, err := os.ReadFile(yamlFile)
// 	if err != nil {
// 		return false, fmt.Sprintf("error reading YAML file: %v", err)
// 	}

// 	var yamlObj interface{}
// 	err = yaml.Unmarshal(yamlData, &yamlObj)
// 	if err != nil {
// 		return false, fmt.Sprintf("error unmarshaling YAML data: %v", err)
// 	}

// 	var r cue.Runtime
// 	binst := load.Instances([]string{cueFile}, &load.Config{})
// 	if len(binst) != 1 || binst[0].Err != nil {
// 		return false, fmt.Sprintf("error loading CUE file: %v", binst[0].Err)
// 	}

// 	cueInstance, err := r.Build(binst[0])
// 	if err != nil {
// 		return false, fmt.Sprintf("error building CUE instance: %v", err)
// 	}

// 	cueValue := cueInstance.Value()
// 	err = cueValue.Validate(cue.Concrete(true))
// 	if err != nil {
// 		return false, fmt.Sprintf("error validating CUE value: %v", err)
// 	}

// 	jsonData, err := json.Marshal(yamlObj)
// 	if err != nil {
// 		return false, fmt.Sprintf("error marshaling YAML object to JSON: %v", err)
// 	}

// 	yamlCueValue, err := r.Compile("", string(jsonData))
// 	if err != nil {
// 		return false, fmt.Sprintf("error compiling JSON data to CUE value: %v", err)
// 	}

// 	err = cueValue.Unify(yamlCueValue.Value()).Validate(cue.Concrete(true))
// 	if err != nil {
// 		return false, fmt.Sprintf("error validating YAML data against CUE schema: %v", err)
// 	}

// 	return true, ""
// }

func validateYAMLAndCUEContent(yamlContent string, cueContent string) (bool, string) {
	var yamlObj interface{}
	err := yaml.Unmarshal([]byte(yamlContent), &yamlObj)
	if err != nil {
		return false, fmt.Sprintf("error unmarshaling YAML data: %v", err)
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

	jsonData, err := json.Marshal(yamlObj)
	if err != nil {
		return false, fmt.Sprintf("error marshaling YAML object to JSON: %v", err)
	}

	yamlCueValue, err := r.Compile("", string(jsonData))
	if err != nil {
		return false, fmt.Sprintf("error compiling JSON data to CUE value: %v", err)
	}

	err = cueValue.Unify(yamlCueValue.Value()).Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating YAML data against CUE schema: %v", err)
	}

	return true, ""
}
