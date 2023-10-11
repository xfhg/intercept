package cmd

import (
	"encoding/json"
	"fmt"

	"cuelang.org/go/cue"
	xdiff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

func validateJSONAndCUEContent(jsonContent string, cueContent string) (bool, string) {
	var jsonObj interface{}
	err := json.Unmarshal([]byte(jsonContent), &jsonObj)
	if err != nil {
		return false, fmt.Sprintf("error unmarshaling json data: %v", err)
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

	jsonData, err := json.Marshal(jsonObj)
	if err != nil {
		return false, fmt.Sprintf("error marshaling json object to JSON: %v", err)
	}

	jsonCueValue, err := r.Compile("", string(jsonData))
	if err != nil {
		return false, fmt.Sprintf("error compiling JSON data to CUE value: %v", err)
	}

	err = cueValue.Unify(jsonCueValue.Value()).Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating json data against CUE schema: %v", err)
	}

	cuepolicy, err := cueValue.Value().MarshalJSON()
	jsoncontent, err := jsonCueValue.Value().MarshalJSON()

	err = cueValue.UnifyAccept(cueValue.Value(), jsonCueValue.Value()).Validate(cue.Concrete(true))
	if err != nil {
		return false, fmt.Sprintf("error validating json accept data against CUE schema: %v", err)
	}

	// subset

	var a, b map[string]interface{}

	json.Unmarshal(cuepolicy, &a)
	json.Unmarshal(jsoncontent, &b)

	if isSubsetOrEqual(a, b) {
		return true, ""
	}

	// lazy match

	keysExist := LazyMatch(b, a)

	// diff

	differ := xdiff.New()
	d, err := differ.Compare(cuepolicy, jsoncontent)
	if err != nil {
		return false, fmt.Sprintf("error unmarshaling content: %s\n", err.Error())
	}

	if d.Modified() && !keysExist {

		var diffString string

		var aJson map[string]interface{}
		json.Unmarshal(cuepolicy, &aJson)

		config := formatter.AsciiFormatterConfig{
			ShowArrayIndex: true,
			Coloring:       true,
		}

		zformatter := formatter.NewAsciiFormatter(aJson, config)
		diffString, err = zformatter.Format(d)
		if err != nil {
			return false, fmt.Sprintf("Internal error: %v", err)
		}

		fmt.Println(diffString)

		return false, fmt.Sprintf("Missing required keys \n")

	} else {
		return true, ""
	}

}
