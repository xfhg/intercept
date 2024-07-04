package cmd

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	xdiff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

func validateJSONAndCUEContent(jsonContent, cueContent string) (bool, string) {
	var jsonObj interface{}
	if err := json.Unmarshal([]byte(jsonContent), &jsonObj); err != nil {
		return false, fmt.Errorf("error unmarshaling JSON data: %w", err).Error()
	}

	ctx := cuecontext.New()
	cueValue := ctx.CompileString(cueContent)
	if cueValue.Err() != nil {
		return false, fmt.Errorf("error compiling CUE content: %w", cueValue.Err()).Error()
	}

	if err := cueValue.Validate(cue.Concrete(true)); err != nil {
		return false, fmt.Errorf("error validating CUE value: %w", err).Error()
	}

	jsonCueValue := ctx.CompileString(jsonContent)
	if jsonCueValue.Err() != nil {
		return false, fmt.Errorf("error compiling JSON data to CUE value: %w", jsonCueValue.Err()).Error()
	}

	if err := cueValue.Unify(jsonCueValue).Validate(cue.Concrete(true)); err != nil {
		return false, fmt.Errorf("error validating JSON data against CUE schema: %w", err).Error()
	}

	cuepolicy, err := cueValue.MarshalJSON()
	if err != nil {
		return false, fmt.Errorf("error marshaling CUE policy to JSON: %w", err).Error()
	}

	jsoncontent, err := jsonCueValue.MarshalJSON()
	if err != nil {
		return false, fmt.Errorf("error marshaling JSON content to CUE JSON: %w", err).Error()
	}

	var a, b map[string]interface{}
	if err := json.Unmarshal(cuepolicy, &a); err != nil {
		return false, fmt.Errorf("error unmarshaling CUE policy: %w", err).Error()
	}
	if err := json.Unmarshal(jsoncontent, &b); err != nil {
		return false, fmt.Errorf("error unmarshaling JSON content: %w", err).Error()
	}

	var cuekeys, allkeys []string
	collectKeys(b, "", &allkeys)
	alloutput := map[string][]string{"keys": allkeys}
	alloutputJSON, err := json.Marshal(alloutput)
	if err != nil {
		return false, fmt.Errorf("error marshaling all keys: %w", err).Error()
	}
	collectKeys(a, "", &cuekeys)
	cueoutput := map[string][]string{"keys": cuekeys}
	cueoutputJSON, err := json.Marshal(cueoutput)
	if err != nil {
		return false, fmt.Errorf("error marshaling CUE keys: %w", err).Error()
	}

	var keysA, keysB KeyArray
	if err := json.Unmarshal(cueoutputJSON, &keysA); err != nil {
		return false, fmt.Errorf("error unmarshaling CUE keys: %w", err).Error()
	}
	if err := json.Unmarshal(alloutputJSON, &keysB); err != nil {
		return false, fmt.Errorf("error unmarshaling all keys: %w", err).Error()
	}
	keysExist := allKeysExist(keysA.Keys, keysB.Keys)

	differ := xdiff.New()
	d, err := differ.Compare(cuepolicy, jsoncontent)
	if err != nil {
		return false, fmt.Errorf("error comparing CUE and JSON content: %w", err).Error()
	}

	if d.Modified() && !keysExist {
		config := formatter.AsciiFormatterConfig{
			ShowArrayIndex: true,
			Coloring:       true,
		}

		formatter := formatter.NewAsciiFormatter(a, config)
		diffString, err := formatter.Format(d)
		if err != nil {
			return false, fmt.Errorf("error formatting diff: %w", err).Error()
		}

		fmt.Println(diffString)
		return false, "Missing required keys"
	}

	return true, ""
}
