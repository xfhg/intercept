package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/lens-vm/jsonmerge"
	"github.com/owenrumney/go-sarif/v2/sarif"
)

type InterceptResult struct {
	Data struct {
		AbsoluteOffset int `json:"absolute_offset"`
		LineNumber     int `json:"line_number"`
		Lines          struct {
			Text string `json:"text"`
		} `json:"lines"`
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Submatches []struct {
			End   int `json:"end"`
			Match struct {
				Text string `json:"text"`
			} `json:"match"`
			Start int `json:"start"`
		} `json:"submatches"`
	} `json:"data"`
	RuleDescription string `json:"ruleDescription"`
	RuleError       string `json:"ruleError"`
	RuleFatal       bool   `json:"ruleFatal"`
	RuleID          string `json:"ruleId"`
	RuleName        string `json:"ruleName"`
	RuleSolution    string `json:"ruleSolution"`
	Type            string `json:"type"`
}

type InterceptOutput []InterceptResult

func ProcessOutput(filename string, ruleId string, ruleName string, ruleDescription string, ruleError string, ruleSolution string, ruleFatal bool) {

	ruleMetaData := map[string]interface{}{
		"ruleId":          ruleId,
		"ruleName":        ruleName,
		"ruleDescription": ruleDescription,
		"ruleError":       ruleError,
		"ruleSolution":    ruleSolution,
		"ruleFatal":       ruleFatal,
	}

	ruleMetajsonData, err := json.Marshal(ruleMetaData)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Open the input file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Read the contents of the file into a string
	content, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Split the string into individual JSON objects
	objects := strings.Split(string(content), "\n")

	// Remove any empty strings from the slice
	var cleanedObjects []string
	for _, object := range objects {
		if object != "" {
			cleanedObjects = append(cleanedObjects, object)
		}
	}

	// Create a new slice for the combined JSON array
	var jsonArray []interface{}

	// Parse each JSON object and add it to the array
	for _, object := range cleanedObjects {
		var jsonObject interface{}
		err := json.Unmarshal([]byte(object), &jsonObject)
		if err != nil {
			fmt.Println(err)
			return
		}
		jsonArray = append(jsonArray, jsonObject)
	}

	// Encode the combined JSON array as a string
	output, err := json.Marshal(jsonArray)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Write the output to a new file
	err = ioutil.WriteFile(filename, []byte(string(output)), 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	////////////////////////////////////////////////////////////////

	var outfile *os.File

	if FileExists("intercept-output.json") {
		outfile, err = os.OpenFile("intercept-output.json", os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
	} else {
		outfile, err = os.Create("intercept-output.json")
		if err != nil {
			panic(err)
		}
	}

	defer outfile.Close()

	// Get the file size
	fileInfo, err := outfile.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()

	if fileSize == 0 {
		// File is empty
		// fmt.Println("File is empty")

		emptyArray := []interface{}{}

		// Marshal the empty slice to JSON
		emptyJSON, err := json.Marshal(emptyArray)
		if err != nil {
			panic(err)
		}

		// Write the JSON to a file
		err = ioutil.WriteFile("intercept-output.json", emptyJSON, 0644)
		if err != nil {
			panic(err)
		}

	}

	var finalobjects []interface{}

	finalerr := json.NewDecoder(outfile).Decode(&finalobjects)
	if finalerr != nil {
		panic(finalerr)
	}
	// fmt.Printf("Found %d JSON objects in file\n", len(finalobjects))

	////////////////////////////////////////////////////////////////

	query, err := gojq.Parse(" .[] | select(.type == \"match\") ")
	if err != nil {
		log.Fatalln(err)
	}

	var filteredJsonArray []interface{}

	iter := query.Run(jsonArray) // or query.RunWithContext
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}

		structured, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}

		new, err := jsonmerge.MergePatch(structured, ruleMetajsonData)
		if err != nil {
			panic(err)
		}

		var newV interface{}
		newerr := json.Unmarshal([]byte(new), &newV)
		if newerr != nil {
			panic(newerr)
		}

		filteredJsonArray = append(filteredJsonArray, newV)

		finalobjects = append(finalobjects, newV)

	}
	finaloutput, ferr := json.Marshal(filteredJsonArray)
	if ferr != nil {
		fmt.Println(ferr)
		return
	}
	err = ioutil.WriteFile(filename, []byte(string(finaloutput)), 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

	////////////////////////////////////////////////////////////////

	compiledoutput, ferr := json.Marshal(finalobjects)
	if ferr != nil {
		fmt.Println(ferr)
		return
	}
	err = ioutil.WriteFile("intercept-output.json", []byte(string(compiledoutput)), 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

}

func GenerateSarif() {
	interceptResults, err := loadInterceptResults()
	if err != nil {
		panic(err)
	}

	// create a new report object
	report, err := sarif.New(sarif.Version210)
	if err != nil {
		panic(err)
	}

	// create a run
	run := sarif.NewRunWithInformationURI("intercept", "https://intercept.cc")

	// for each result, add the
	for _, r := range interceptResults {

		// create a property bag for the non standard stuff
		pb := sarif.NewPropertyBag()
		pb.Add("impact", r.RuleError)
		pb.Add("resolution", r.RuleSolution)

		// create a new rule for each rule id
		run.AddRule(strings.Join([]string{"intercept.cc.policy.", r.RuleID, ": ", r.RuleName}, "")).
			WithDescription(r.RuleDescription).
			WithHelpURI("https://intercept.cc").
			WithProperties(pb.Properties).
			WithMarkdownHelp("# INTERCEPT")

		// add the location as a unique artifact
		run.AddDistinctArtifact(r.Data.Path.Text)

		ResultLevel := func() string {
			if r.RuleFatal {
				return "error"
			}
			return "warning"
		}()

		snippetText := strings.Trim(r.Data.Submatches[0].Match.Text, "\n")

		artifactContent := sarif.ArtifactContent{
			Text: &snippetText,
		}

		// add each of the results with the details of where the issue occurred
		run.CreateResultForRule(strings.Join([]string{"intercept.cc.policy.", r.RuleID, ": ", r.RuleName}, "")).
			WithLevel(strings.ToLower(ResultLevel)).
			WithMessage(sarif.NewTextMessage(r.RuleDescription)).
			AddLocation(
				sarif.NewLocationWithPhysicalLocation(
					sarif.NewPhysicalLocation().
						WithArtifactLocation(
							sarif.NewSimpleArtifactLocation(r.Data.Path.Text),
						).WithRegion(
						sarif.NewSimpleRegion(r.Data.LineNumber, r.Data.LineNumber).WithSnippet(&artifactContent),
					),
				),
			)
	}

	report.AddRun(run)

	// print the report to stdout
	//_ = report.PrettyWrite(os.Stdout)

	if err := report.WriteFile("intercept-sarif.json"); err != nil {
		panic(err)
	}

}

func loadInterceptResults() (InterceptOutput, error) {

	jsonResult, err := ioutil.ReadFile("intercept-output.json")
	if err != nil {
		panic(err)
	}

	var results InterceptOutput

	err = json.Unmarshal(jsonResult, &results)
	return results, err
}
