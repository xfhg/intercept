package cmd

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/itchyny/gojq"
	"github.com/lens-vm/jsonmerge"
	"github.com/owenrumney/go-sarif/v2/sarif"
)

type SARIF struct {
	Version string        `json:"version"`
	Schema  string        `json:"$schema"`
	Runs    []interface{} `json:"runs"`
}

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
	RuleType        string `json:"ruleType"`
	Type            string `json:"type"`
}
type InterceptOutput []InterceptResult

type InterceptComplianceFinding struct {
	FileName  string `yaml:"filename"`
	FileHash  string `yaml:"filehash"`
	Output    string `yaml:"output"`
	Compliant bool   `yaml:"compliant"`
	Missing   bool   `yaml:"missing"`
	ParentID  string `yaml:"parentID"`
}

type InterceptCompliance struct {
	RuleFindings    []InterceptComplianceFinding `json:"ruleFindings"`
	RuleDescription string                       `json:"ruleDescription"`
	RuleError       string                       `json:"ruleError"`
	RuleFatal       bool                         `json:"ruleFatal"`
	RuleID          string                       `json:"ruleId"`
	RuleName        string                       `json:"ruleName"`
	RuleSolution    string                       `json:"ruleSolution"`
	RuleType        string                       `json:"ruleType"`
	Type            string                       `json:"type"`
}

type InterceptComplianceOutput []InterceptCompliance

// processes ripgrep output into intercept meaningfull results
func ProcessOutput(filename string, ruleId string, ruleType string, ruleName string, ruleDescription string, ruleError string, ruleSolution string, ruleFatal bool) {

	ruleMetaData := map[string]interface{}{
		"ruleId":          ruleId,
		"ruleName":        ruleName,
		"ruleDescription": ruleDescription,
		"ruleError":       ruleError,
		"ruleSolution":    ruleSolution,
		"ruleFatal":       ruleFatal,
		"ruleType":        ruleType,
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

	if FileExists("intercept.audit.output.json") {
		outfile, err = os.OpenFile("intercept.audit.output.json", os.O_RDWR, 0644)
		if err != nil {
			LogError(err)
		}
	} else {
		outfile, err = os.Create("intercept.audit.output.json")
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
		err = os.WriteFile("intercept.audit.output.json", emptyJSON, 0644)
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
	err = os.WriteFile("intercept.audit.output.json", []byte(string(compiledoutput)), 0644)
	if err != nil {
		LogError(err)
		return
	}

	if FileExists(filename) {
		os.Remove(filename)
	}

}

// package rg output into intercept results struct for SARIF
func loadInterceptResults() (InterceptOutput, error) {

	if FileExists("intercept.audit.output.json") {

		jsonResult, err := os.ReadFile("intercept.audit.output.json")
		if err != nil {
			LogError(err)
		}

		var results InterceptOutput

		err = json.Unmarshal(jsonResult, &results)

		return results, err

	} else {
		return nil, errors.New("no results found")
	}
}

// generates SARIF output from rg intercept results
func GenerateSarif(calledby string) {

	// input needs cmd ran

	interceptResults, err := loadInterceptResults()
	if err != nil {
		// clean scan return nothing
		return
	}

	report, err := sarif.New(sarif.Version210)
	if err != nil {
		LogError(err)
	}

	// build strings

	run := sarif.NewRunWithInformationURI("intercept", "https://intercept.cc")

	if buildVersion != "" {
		run.Tool.Driver.SemanticVersion = &buildVersion
	}

	for _, r := range interceptResults {

		r.RuleType = strings.ToLower(r.RuleType)

		pb := sarif.NewPropertyBag()
		pb.Add("impact", r.RuleError)
		pb.Add("resolution", r.RuleSolution)

		run.AddRule(strings.Join([]string{"intercept.cc.", strings.ToLower(r.RuleType), ".policy.", r.RuleID, ": ", strings.ToUpper(r.RuleName)}, "")).
			WithDescription(r.RuleDescription).
			WithHelpURI("https://intercept.cc").
			WithProperties(pb.Properties).
			WithMarkdownHelp(ToPascalCase(strings.ToUpper(r.RuleName))).WithTextHelp(r.RuleSolution)

		run.AddDistinctArtifact(r.Data.Path.Text)

		ResultLevel := func() string {
			if r.RuleType == "api" && r.Type == "match" {
				return "note"
			}
			if r.RuleType == "collect" || r.RuleType == "assure" {
				return "note"
			} else {
				if r.RuleFatal {
					return "error"
				}
				return "warning"
			}

		}()

		snippetText := strings.Trim(r.Data.Submatches[0].Match.Text, "\n")

		artifactContent := sarif.ArtifactContent{
			Text: &snippetText,
		}

		run.CreateResultForRule(strings.Join([]string{"intercept.cc.", strings.ToLower(r.RuleType), ".policy.", r.RuleID, ": ", strings.ToUpper(r.RuleName)}, "")).
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

	if err := report.WriteFile("intercept.audit.sarif.json"); err != nil {
		LogError(err)
	}

}

func FileToBase64(filepath string) string {
	file, err := os.Open(filepath)
	if err != nil {
		LogError(err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	fileBytes := make([]byte, size)

	buffer := bufio.NewReader(file)
	_, err = buffer.Read(fileBytes)

	if err != nil {
		LogError(err)
	}

	fileBase64 := base64.StdEncoding.EncodeToString(fileBytes)
	return fileBase64
}

func GenerateComplianceSarif(results InterceptComplianceOutput) {

	report, err := sarif.New(sarif.Version210)
	if err != nil {
		LogError(err)
	}

	// build strings

	findings := 0

	run := sarif.NewRunWithInformationURI("intercept", "https://intercept.cc")

	if buildVersion != "" {
		run.Tool.Driver.SemanticVersion = &buildVersion
	}

	for _, r := range results {

		r.RuleType = strings.ToLower(r.RuleType)

		pb := sarif.NewPropertyBag()
		pb.Add("impact", r.RuleError)
		pb.Add("resolution", r.RuleSolution)

		run.AddRule(strings.Join([]string{"intercept.cc.", strings.ToLower(r.RuleType), ".policy.", r.RuleID, ": ", strings.ToUpper(r.RuleName)}, "")).
			WithDescription(r.RuleDescription).
			WithHelpURI("https://intercept.cc").
			WithProperties(pb.Properties).
			WithMarkdownHelp(ToPascalCase(strings.ToUpper(r.RuleName))).WithTextHelp(r.RuleSolution)

		for _, rf := range r.RuleFindings {

			findings++

			run.AddDistinctArtifact(rf.FileHash)

			ResultLevel := func() string {
				if rf.Compliant {
					return "note"
				} else {
					if rf.Missing {
						return "warning"
					}
					return "error"
				}

			}()

			snippetText := strings.Trim(rf.Output, "\n")

			artifactContent := sarif.ArtifactContent{
				Text: &snippetText,
			}

			run.CreateResultForRule(strings.Join([]string{"intercept.cc.", strings.ToLower(r.RuleType), ".policy.", r.RuleID, ": ", strings.ToUpper(r.RuleName)}, "")).
				WithLevel(strings.ToLower(ResultLevel)).
				WithMessage(sarif.NewTextMessage(r.RuleDescription)).
				AddLocation(
					sarif.NewLocationWithPhysicalLocation(
						sarif.NewPhysicalLocation().
							WithArtifactLocation(
								sarif.NewSimpleArtifactLocation(rf.FileName),
							).WithRegion(
							sarif.NewSimpleRegion(1, 1).WithSnippet(&artifactContent),
						),
					),
				)
		}

	}

	report.AddRun(run)

	sarifOutputFilename := strings.Join([]string{"intercept.", strings.ToLower(results[0].RuleType), ".", strconv.FormatInt(time.Now().UnixNano(), 10), ".sarif.json"}, "")
	// Commenting existing check to see if the file exists and the subsequent removal as it is not required for unique file names
	/*
		if FileExists(sarifOutputFilename) {
			_ = os.Remove(sarifOutputFilename)
		}
	*/
	if findings > 0 {
		if err := report.WriteFile(sarifOutputFilename); err != nil {
			LogError(err)
		}
	}

}

func MergeSARIFFiles(file1, file2 string) (*SARIF, error) {
	// Read and unmarshal the first file
	data1, err := os.ReadFile(file1)
	if err != nil {
		return nil, err
	}
	var sarif1 SARIF
	if err := json.Unmarshal(data1, &sarif1); err != nil {
		return nil, err
	}

	// Read and unmarshal the second file
	data2, err := os.ReadFile(file2)
	if err != nil {
		return nil, err
	}
	var sarif2 SARIF
	if err := json.Unmarshal(data2, &sarif2); err != nil {
		return nil, err
	}

	// Merge the 'runs' slices
	mergedRuns := append(sarif1.Runs, sarif2.Runs...)

	// Create the merged SARIF
	mergedSARIF := &SARIF{
		Version: sarif1.Version,
		Schema:  sarif1.Schema,
		Runs:    mergedRuns,
	}

	return mergedSARIF, nil
}

func GenerateApiSARIF() {

	if FileExists("intercept.audit.sarif.json") && FileExists("intercept.api.sarif.json") {

		mergedSARIF, err := MergeSARIFFiles("intercept.audit.sarif.json", "intercept.api.sarif.json")
		if err != nil {
			LogError(err)
		}

		mergedData, err := json.MarshalIndent(mergedSARIF, "", "    ")
		if err != nil {
			LogError(err)
		}

		if err := os.WriteFile("intercept.api.full.sarif.json", mergedData, 0644); err != nil {
			LogError(err)
		}

	} else {

		auditdata, err := os.ReadFile("intercept.audit.sarif.json")
		if err != nil {
			LogError(err)
		}

		if err := os.WriteFile("intercept.api.full.sarif.json", auditdata, 0644); err != nil {
			LogError(err)
		}
	}

}

func GenerateAssureSARIF() {

	if FileExists("intercept.audit.sarif.json") && FileExists("intercept.assure.sarif.json") {

		mergedSARIF, err := MergeSARIFFiles("intercept.audit.sarif.json", "intercept.assure.sarif.json")
		if err != nil {
			LogError(err)
		}

		mergedData, err := json.MarshalIndent(mergedSARIF, "", "    ")
		if err != nil {
			LogError(err)
		}

		if err := os.WriteFile("intercept.assure.full.sarif.json", mergedData, 0644); err != nil {
			LogError(err)
		}

	} else {

		auditdata, err := os.ReadFile("intercept.audit.sarif.json")
		if err != nil {
			LogError(err)
		}

		if err := os.WriteFile("intercept.assure.full.sarif.json", auditdata, 0644); err != nil {
			LogError(err)
		}
	}

}

func readExternalData(filePath string) (map[string]interface{}, error) {
	var data map[string]interface{}

	// Read the file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a map
	if err := json.Unmarshal(fileData, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func ToPascalCase(s string) string {
	// Remove non-alphanumeric characters
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	s = re.ReplaceAllString(s, "")

	// Split the string into words
	words := strings.Fields(s)

	// Capitalize the first letter of each word and join them
	var pascalCase string
	for _, word := range words {
		if len(word) > 0 {
			pascalCase += strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}

	return pascalCase
}
