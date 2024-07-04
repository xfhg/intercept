package cmd

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	xdiff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"gopkg.in/yaml.v3"
)

func validateYAMLAndCUEContent(yamlContent, cueContent string) (bool, string) {
	var yamlObj interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &yamlObj); err != nil {
		return false, fmt.Sprintf("error unmarshaling YAML data: %v", err)
	}

	ctx := cuecontext.New()
	cueValue := ctx.CompileString(cueContent)
	if cueValue.Err() != nil {
		return false, fmt.Sprintf("error compiling CUE content: %v", cueValue.Err())
	}

	if err := cueValue.Validate(cue.Concrete(true)); err != nil {
		return false, fmt.Sprintf("error validating CUE value: %v", err)
	}

	jsonData, err := json.Marshal(yamlObj)
	if err != nil {
		return false, fmt.Sprintf("error marshaling YAML object to JSON: %v", err)
	}

	yamlCueValue := ctx.CompileBytes(jsonData)
	if yamlCueValue.Err() != nil {
		return false, fmt.Sprintf("error compiling JSON data to CUE value: %v", yamlCueValue.Err())
	}

	cuepolicy, err := cueValue.MarshalJSON()
	if err != nil {
		return false, fmt.Sprintf("error marshaling CUE policy to JSON: %v", err)
	}

	yamlcontent, err := yamlCueValue.MarshalJSON()
	if err != nil {
		return false, fmt.Sprintf("error marshaling YAML content to JSON: %v", err)
	}

	if err := cueValue.Unify(yamlCueValue).Validate(cue.Concrete(true)); err != nil {
		return false, fmt.Sprintf("error validating YAML data against CUE schema: %v", err)
	}

	var a, b map[string]interface{}
	if err := json.Unmarshal(cuepolicy, &a); err != nil {
		return false, fmt.Sprintf("error unmarshaling CUE policy: %v", err)
	}
	if err := json.Unmarshal(yamlcontent, &b); err != nil {
		return false, fmt.Sprintf("error unmarshaling YAML content: %v", err)
	}

	var cuekeys, allkeys []string
	collectKeys(b, "", &allkeys)
	alloutput := map[string][]string{"keys": allkeys}
	alloutputJSON, _ := json.Marshal(alloutput)
	collectKeys(a, "", &cuekeys)
	cueoutput := map[string][]string{"keys": cuekeys}
	cueoutputJSON, _ := json.Marshal(cueoutput)

	var keysA, keysB KeyArray
	if err := json.Unmarshal(cueoutputJSON, &keysA); err != nil {
		return false, fmt.Sprintf("error unmarshaling CUE keys: %v", err)
	}
	if err := json.Unmarshal(alloutputJSON, &keysB); err != nil {
		return false, fmt.Sprintf("error unmarshaling all keys: %v", err)
	}
	keysExist := allKeysExist(keysA.Keys, keysB.Keys)

	differ := xdiff.New()
	d, err := differ.Compare(cuepolicy, yamlcontent)
	if err != nil {
		return false, fmt.Sprintf("error comparing CUE and YAML content: %v", err)
	}

	if d.Modified() && !keysExist {
		config := formatter.AsciiFormatterConfig{
			ShowArrayIndex: true,
			Coloring:       true,
		}

		formatter := formatter.NewAsciiFormatter(a, config)
		diffString, err := formatter.Format(d)
		if err != nil {
			return false, fmt.Sprintf("error formatting diff: %v", err)
		}

		fmt.Println(diffString)
		return false, "Missing required keys"
	}

	return true, ""
}
