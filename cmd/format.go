package cmd

import (
	"encoding/json"
	"io"
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
		LogError(err)
		return
	}
	file, err := os.Open(filename)
	if err != nil {
		LogError(err)
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		LogError(err)
		return
	}

	objects := strings.Split(string(content), "\n")

	var cleanedObjects []string
	for _, object := range objects {
		if object != "" {
			cleanedObjects = append(cleanedObjects, object)
		}
	}

	var jsonArray []interface{}

	for _, object := range cleanedObjects {
		var jsonObject interface{}
		err := json.Unmarshal([]byte(object), &jsonObject)
		if err != nil {
			LogError(err)
			return
		}
		jsonArray = append(jsonArray, jsonObject)
	}

	output, err := json.Marshal(jsonArray)
	if err != nil {
		LogError(err)
		return
	}

	err = os.WriteFile(filename, []byte(string(output)), 0644)
	if err != nil {
		LogError(err)
		return
	}

	////////////////////////////////////////////////////////////////

	var outfile *os.File

	if FileExists("intercept.output.json") {
		outfile, err = os.OpenFile("intercept.output.json", os.O_RDWR, 0644)
		if err != nil {
			LogError(err)
		}
	} else {
		outfile, err = os.Create("intercept.output.json")
		if err != nil {
			LogError(err)
		}
	}

	defer outfile.Close()

	fileInfo, err := outfile.Stat()
	if err != nil {
		LogError(err)
	}
	fileSize := fileInfo.Size()

	if fileSize == 0 {

		emptyArray := []interface{}{}

		emptyJSON, err := json.Marshal(emptyArray)
		if err != nil {
			LogError(err)
		}

		// Write the JSON to a file
		err = os.WriteFile("intercept.output.json", emptyJSON, 0644)
		if err != nil {
			LogError(err)
		}

	}

	var finalobjects []interface{}

	finalerr := json.NewDecoder(outfile).Decode(&finalobjects)
	if finalerr != nil {
		LogError(finalerr)
	}

	query, err := gojq.Parse(" .[] | select(.type == \"match\") ")
	if err != nil {
		LogError(err)
	}

	var filteredJsonArray []interface{}

	iter := query.Run(jsonArray)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			LogError(err)
		}

		structured, err := json.Marshal(v)
		if err != nil {
			LogError(err)
		}

		new, err := jsonmerge.MergePatch(structured, ruleMetajsonData)
		if err != nil {
			LogError(err)
		}

		var newV interface{}
		newerr := json.Unmarshal([]byte(new), &newV)
		if newerr != nil {
			LogError(newerr)
		}

		filteredJsonArray = append(filteredJsonArray, newV)

		finalobjects = append(finalobjects, newV)

	}
	finaloutput, ferr := json.Marshal(filteredJsonArray)
	if ferr != nil {
		LogError(ferr)
		return
	}
	err = os.WriteFile(filename, []byte(string(finaloutput)), 0644)
	if err != nil {
		LogError(err)
		return
	}

	compiledoutput, ferr := json.Marshal(finalobjects)
	if ferr != nil {
		LogError(ferr)
		return
	}
	err = os.WriteFile("intercept.output.json", []byte(string(compiledoutput)), 0644)
	if err != nil {
		LogError(err)
		return
	}

}

func GenerateSarif() {
	interceptResults, err := loadInterceptResults()
	if err != nil {
		LogError(err)
	}

	report, err := sarif.New(sarif.Version210)
	if err != nil {
		LogError(err)
	}

	run := sarif.NewRunWithInformationURI("intercept", "https://intercept.cc")

	for _, r := range interceptResults {

		pb := sarif.NewPropertyBag()
		pb.Add("impact", r.RuleError)
		pb.Add("resolution", r.RuleSolution)

		run.AddRule(strings.Join([]string{"intercept.cc.policy.", r.RuleID, ": ", r.RuleName}, "")).
			WithDescription(r.RuleDescription).
			WithHelpURI("https://intercept.cc").
			WithProperties(pb.Properties).
			WithMarkdownHelp("# INTERCEPT")

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

	if err := report.WriteFile("intercept.sarif.json"); err != nil {
		LogError(err)
	}

}

func loadInterceptResults() (InterceptOutput, error) {

	jsonResult, err := os.ReadFile("intercept.output.json")
	if err != nil {
		LogError(err)
	}

	var results InterceptOutput

	err = json.Unmarshal(jsonResult, &results)
	return results, err
}
