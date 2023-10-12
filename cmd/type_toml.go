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

	cuepolicy, err := cueValue.Value().MarshalJSON()
	tomlcontent, err := tomlCueValue.Value().MarshalJSON()

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
				fmt.Printf("Key '%s' is absent\n", key)
				return false, fmt.Sprintf("Key '%s' is absent", key)
			}
		}
	}

	// subset

	// var a, b map[string]interface{}

	// json.Unmarshal(cuepolicy, &a)
	// json.Unmarshal(tomlcontent, &b)

	// if isSubsetOrEqual(a, b) {
	// 	return true, ""
	// }

	// lazy match

	// keysExist := LazyMatch(b, a)

	// lazy match

	// keysExist := LazyMatch(b, a)

	// diff

	// differ := xdiff.New()
	// d, err := differ.Compare(cuepolicy, tomlcontent)
	// if err != nil {
	// 	return false, fmt.Sprintf("error unmarshaling content: %s\n", err.Error())
	// }

	// if d.Modified() && !keysExist {

	// 	var diffString string

	// 	var aJson map[string]interface{}
	// 	json.Unmarshal(cuepolicy, &aJson)

	// 	config := formatter.AsciiFormatterConfig{
	// 		ShowArrayIndex: true,
	// 		Coloring:       true,
	// 	}

	// 	zformatter := formatter.NewAsciiFormatter(aJson, config)
	// 	diffString, err = zformatter.Format(d)
	// 	if err != nil {
	// 		return false, fmt.Sprintf("Internal error: %v", err)
	// 	}

	// 	fmt.Println(diffString)

	// 	return false, fmt.Sprintf("Missing required keys \n")

	// } else {
	// 	return true, ""
	// }

	return true, ""
}
