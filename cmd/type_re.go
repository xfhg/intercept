package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage/inmem"
	"github.com/open-policy-agent/opa/topdown"
)

var (
	rCompliance InterceptComplianceOutput
	rRule       InterceptCompliance
)

func processRegoType(value Rule) {

	//exception := ContainsInt(rules.Exceptions, value.ID)

	//if exception && !auditNox && !value.Enforcement {
	if !auditNox && !value.Enforcement {

		colorRedBold.Println("│")
		colorRedBold.Println("│ ", rules.ExceptionMessage)
		colorRedBold.Println("│")

	} else {

		// REGO

		if value.Rego_Filepattern != "" && value.Rego_Policy_File != "" && value.Rego_Policy_Query != "" {

			fmt.Println("│ ")
			fmt.Println(line)
			fmt.Println("│ ")
			fmt.Println("├ REGO Rule #", value.ID)
			fmt.Println("│ Rule name : ", value.Name)
			fmt.Println("│ Rule description : ", value.Description)
			fmt.Println("│ Impacted Env : ", value.Environment)
			fmt.Println("│ Confidence : ", value.Confidence)
			fmt.Println("│ Tags : ", value.Tags)
			fmt.Println("│ REGO Input File Pattern : ", value.Rego_Filepattern)
			fmt.Println("│ REGO Policy : ", value.Rego_Policy_File)
			fmt.Println("│ REGO Query : ", value.Rego_Policy_Query)
			fmt.Println("│ ")

			// SARIF Structure for REGO
			rRule = InterceptCompliance{}
			rRule.RuleDescription = value.Description
			rRule.RuleError = value.Error
			rRule.RuleFatal = value.Fatal
			rRule.RuleID = value.ID
			rRule.RuleName = value.Name
			rRule.RuleSolution = value.Solution
			rRule.RuleType = value.Type

			// Load REGO POLICY
			policyBytes, err := os.ReadFile(value.Rego_Policy_File)
			if err != nil {
				LogError(fmt.Errorf("error reading REGO policy file: %s", err))
				return
			}
			policyRules := string(policyBytes)

			// Compile regex pattern
			regex, err := regexp.Compile(value.Rego_Filepattern)
			if err != nil {
				LogError(fmt.Errorf("error compiling regex: %s", err))
				return
			}

			// IF NOT JSON INPUT/TARGET , try to converto to JSON (HCL,YAML,XML)
			// TODO

			fileInfo, err := os.Stat(scanPath)
			if err != nil {
				LogError(fmt.Errorf("error accessing path: %s", err))
			}

			// Check if the path is a directory
			if fileInfo.IsDir() {
				// fmt.Printf("%s is a directory\n", jsonScanPath)
			} else {
				LogError(fmt.Errorf("%s is not a directory", scanPath))
			}

			// Function to match files
			err = filepath.Walk(scanPath,
				func(path string, info os.FileInfo, err error) error {

					if err != nil {
						return err
					}

					rFinding := InterceptComplianceFinding{}

					// Ignore directories and match files against regex pattern
					if !info.IsDir() && regex.MatchString(info.Name()) {

						fmt.Println("│ Scanning..")
						fmt.Println("│ File : " + path) // Print matched file path
						filehash, _ := calculateSHA256(path)
						fmt.Println("│ Hash : " + filehash)
						fmt.Println("│")

						// -------------------------------------------------------------- Tracer

						// Setup a Tracer for debugging
						ctx := context.Background()
						tracer := topdown.NewBufferTracer()

						// -------------------------------------------------------------- Input file

						var input map[string]interface{}

						input, err = readExternalData(path)
						if err != nil {
							fmt.Println("Error reading external data:", err)
							return nil
						}

						// -------------------------------------------------------------- External Data
						r := rego.New()

						if value.Rego_Policy_Data != "" {
							var externalData map[string]interface{}

							externalData, err = readExternalData(value.Rego_Policy_Data)
							if err != nil {
								fmt.Println("Error reading external data:", err)
								return nil
							}

							store := inmem.NewFromObject(externalData)

							// -------------------------------------------------------------- New REGO

							r = rego.New(
								rego.Query(value.Rego_Policy_Query),
								rego.Module(value.Rego_Policy_File, policyRules),
								rego.Input(input),
								rego.Store(store),
								rego.Tracer(tracer),
							)

						} else {
							r = rego.New(
								rego.Query(value.Rego_Policy_Query),
								rego.Module(value.Rego_Policy_File, policyRules),
								rego.Input(input),
								rego.Tracer(tracer),
							)
						}

						// -------------------------------------------------------------- Prepare for Evaluation
						query, err := r.PrepareForEval(ctx)
						if err != nil {
							LogError(fmt.Errorf("error preparing query: %s", err))
							return nil
						}

						// -------------------------------------------------------------- Evaluate
						results, err := query.Eval(ctx, rego.EvalTracer(tracer))
						if err != nil {
							LogError(fmt.Errorf("error evaluating policy: %s", err))
							return nil
						}

						// -------------------------------------------------------------- Print Tracer

						fmt.Println("│ ")
						colorGreenBold.Println("│ REGO Tracer : ")
						fmt.Println("")
						topdown.PrettyTrace(os.Stdout, *tracer)
						fmt.Println("")
						fmt.Println("│ ")

						// -------------------------------------------------------------- Check Compliance

						rFinding = InterceptComplianceFinding{
							FileName: path,
							FileHash: filehash,
							ParentID: value.ID,
						}

						for _, result := range results {
							// Each result can have one or more expressions
							for i, expression := range result.Expressions {

								fmt.Printf("│ Query %d : %s \n", i, expression.Text)
								fmt.Printf("│ Result %d : ", i)
								switch value := expression.Value.(type) {
								case bool:
									fmt.Println(value)
								case float64, int, json.Number:
									fmt.Println(value)
								case string:
									fmt.Println(value)
								case map[string]interface{}:
									for key, val := range value {
										fmt.Printf(" %s : %v \n", key, val)
									}
								default:
									fmt.Println("Unsupported type")
								}

							}
						}
						fmt.Println("│ ")

						if results.Allowed() {

							colorGreenBold.Println("│ ")
							colorGreenBold.Println("│ Compliant")
							colorGreenBold.Println("│ ")
							colorGreenBold.Println("├── ─")
							fmt.Println("│")
							stats.Clean++
							stats.Total++
							rFinding.Output = "COMPLIANT"
							rFinding.Compliant = true
							rFinding.Missing = false

						} else {
							colorRedBold.Println("│")
							colorRedBold.Println("│ NON COMPLIANT : ")
							colorRedBold.Println("│ ", value.Error)
							colorRedBold.Println("│ ", value.Solution)
							colorRedBold.Println("│")
							colorRedBold.Println("├── ─")
							fmt.Println("│")
							stats.Dirty++
							stats.Total++
							rFinding.Output = "NON COMPLIANT"
							rFinding.Compliant = false
							rFinding.Missing = false
							warning = true
							envfound := FindMatchingString(cfgEnv, value.Environment, ",")
							if (envfound || strings.Contains(value.Environment, "all") || value.Environment == "") && value.Fatal {
								fatal = true
								stats.Fatal++
							}
						}

						rRule.RuleFindings = append(rRule.RuleFindings, rFinding)

					}

					return nil

				},
			)
			if err != nil {

				LogError(fmt.Errorf("error walking directory: %s", err))
			}

		} else {

			LogError(errors.New("REGO Required policy fields not set (REGO File Pattern / REGO Policy Data / REGO Policy Query)"))

		}

		rCompliance = append(rCompliance, rRule)

	}

}
