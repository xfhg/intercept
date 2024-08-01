package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var (
	tomlTurboAPI  string
	tomlScanPath  string
	tomlscanTags  string
	tomlscanBreak string
	otCompliance  InterceptComplianceOutput
	otRule        InterceptCompliance
)

var tomlCmd = &cobra.Command{
	Use:   "toml",
	Short: "INTERCEPT / TOML - Scan TOML structure from configured policy rules",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if cfgEnv == "" {
			cfgEnv = "先锋"
		}

		cleanupFiles()

		rules = loadUpRules()

		stats = allStats{0, 0, 0, 0}

		startTime := time.Now()

		formattedTime := startTime.Format("2006-01-02 15:04:05")

		fmt.Println("├ S ", formattedTime)
		fmt.Println("│")

		line := "├────────────────────────────────────────────────────────────"

		fmt.Println("│ ")

		fmt.Println("│ Scan ENV\t: ", cfgEnv)
		fmt.Println("│ Scan TAG\t: ", tomlscanTags)
		fmt.Println("│ ")
		fmt.Println(line)

		fmt.Println("│ ")
		fmt.Println("│ ")
		if rules.Banner != "" {
			fmt.Println(rules.Banner)
		}
		fmt.Println("│ ")
		if len(rules.Rules) < 1 {
			colorYellowBold.Println("│ No Policy rules detected")
			fmt.Println("│ Run the command - intercept config - to setup policies")
			PrintClose()
			os.Exit(0)
		}

		if tomlTurboAPI == "false" && len(rules.Rules) < 50 {

			for _, value := range rules.Rules {

				tagfound := FindMatchingString(tomlscanTags, value.Tags, ",")
				if tagfound || tomlscanTags == "" {
					fmt.Println("│ ")
				} else {
					continue
				}

				switch value.Type {

				case "toml":

					if value.Toml_Filepattern != "" && value.Toml_Structure != "" {

						fmt.Println("│ ")
						fmt.Println(line)
						fmt.Println("│ ")
						fmt.Println("├ TOML Rule #", value.ID)
						fmt.Println("│ Rule name : ", value.Name)
						fmt.Println("│ Rule description : ", value.Description)
						fmt.Println("│ Impacted Env : ", value.Environment)
						fmt.Println("│ Confidence : ", value.Confidence)
						fmt.Println("│ Tags : ", value.Tags)
						fmt.Println("│ File pattern : ", value.Toml_Filepattern)
						fmt.Println("│ TOML pattern : ", value.Toml_Structure)
						fmt.Println("│ ")

						otRule = InterceptCompliance{}
						otRule.RuleDescription = value.Description
						otRule.RuleError = value.Error
						otRule.RuleFatal = value.Fatal
						otRule.RuleID = value.ID
						otRule.RuleName = value.Name
						otRule.RuleSolution = value.Solution
						otRule.RuleType = value.Type

						// Compile regex pattern
						regex, err := regexp.Compile(value.Toml_Filepattern)
						if err != nil {
							LogError(fmt.Errorf("error compiling regex: %s", err))
							return
						}

						fileInfo, err := os.Stat(tomlScanPath)
						if err != nil {
							LogError(fmt.Errorf("error accessing path: %s", err))
						}

						// Check if the path is a directory
						if fileInfo.IsDir() {
							// fmt.Printf("%s is a directory\n", tomlScanPath)
						} else {
							LogError(fmt.Errorf("%s is not a directory", tomlScanPath))
						}

						// Function to match files
						err = filepath.Walk(tomlScanPath, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								return err
							}

							oFinding := InterceptComplianceFinding{}

							// Ignore directories and match files against regex pattern
							if !info.IsDir() && regex.MatchString(info.Name()) {

								fmt.Println("│ Scanning..")
								fmt.Println("│ File : " + path) // Print matched file path
								filehash, _ := calculateSHA256(path)
								fmt.Println("│ Hash : " + filehash)
								fmt.Println("│")

								// Read the file in path to a string ymlContent

								ymlContentBytes, err := os.ReadFile(path)
								if err != nil {
									LogError(fmt.Errorf("error reading file: %s", err))
									return nil
								}

								var tomlObj interface{}
								err = toml.Unmarshal(ymlContentBytes, &tomlObj)
								if err != nil {
									LogError(fmt.Errorf("error unmarshaling toml data: %v", err))
									return nil
								}

								oFinding = InterceptComplianceFinding{
									FileName: path,
									FileHash: filehash,
									ParentID: value.ID,
								}

								ymlContent := string(ymlContentBytes)

								xvalid, xerrMsg := validateTOMLAndCUEContent(ymlContent, value.Toml_Structure)
								if xvalid {
									// fmt.Printf("│ The toml file is valid.\n")
									colorGreenBold.Println("│ ")
									colorGreenBold.Println("│ Compliant")
									colorGreenBold.Println("│ ")
									colorGreenBold.Println("├── ─")
									fmt.Println("│")
									stats.Clean++
									stats.Total++

									oFinding.Output = "COMPLIANT"
									oFinding.Compliant = true
									oFinding.Missing = false

								} else {

									envfound := FindMatchingString(cfgEnv, value.Environment, ",")
									if (envfound || strings.Contains(value.Environment, "all") || value.Environment == "") && value.Fatal {
										fatal = true
										stats.Fatal++
									}

									// fmt.Printf("│ The toml file is not valid: %s\n", xerrMsg)
									colorRedBold.Println("│")
									colorRedBold.Println("│ NON COMPLIANT : ")
									colorRedBold.Println("│ ", value.Error)
									colorRedBold.Println("│ ", xerrMsg)
									colorRedBold.Println("│")
									colorRedBold.Println("├── ─")
									fmt.Println("│")
									stats.Dirty++
									stats.Total++
									warning = true

									isMissing := FindMatchingString("Missing", xerrMsg, " ")
									oFinding.Output = xerrMsg
									oFinding.Compliant = false
									oFinding.Missing = false
									if isMissing {
										oFinding.Missing = true
									}

								}
								otRule.RuleFindings = append(otRule.RuleFindings, oFinding)

							}
							return nil
						})
						if err != nil {

							LogError(fmt.Errorf("error walking directory: %s", err))
						}

					} else {

						LogError(errors.New("TOML Required policy fields not set (File Pattern / TOML Structure)"))

					}

					otCompliance = append(otCompliance, otRule)

				default:

				}

			}

		} else {

			var wg sync.WaitGroup

			fmt.Println("├─ Scanning.. ─ ─")

			for _, value := range rules.Rules {

				wg.Add(1) // Increment the WaitGroup counter

				go func(value Rule) { // Launch a goroutine with the loop value
					defer wg.Done() // Decrement the counter when the goroutine completes

					tagfound := FindMatchingString(tomlscanTags, value.Tags, ",")
					if tagfound || tomlscanTags == "" {
						fmt.Println("│ ")
					} else {
						return
					}

					switch value.Type {

					case "toml":

						if value.Toml_Filepattern != "" && value.Toml_Structure != "" {

							regex, err := regexp.Compile(value.Toml_Filepattern)
							if err != nil {
								LogError(fmt.Errorf("error compiling regex: %s", err))
								return
							}

							fileInfo, err := os.Stat(tomlScanPath)
							if err != nil {
								LogError(fmt.Errorf("error accessing path: %s", err))
							}

							// Check if the path is a directory
							if fileInfo.IsDir() {
								// fmt.Printf("%s is a directory\n", tomlScanPath)
							} else {
								LogError(fmt.Errorf("%s is not a directory", tomlScanPath))
							}

							// Function to match files
							err = filepath.Walk(tomlScanPath, func(path string, info os.FileInfo, err error) error {
								if err != nil {
									return err
								}
								// Ignore directories and match files against regex pattern
								if !info.IsDir() && regex.MatchString(info.Name()) {

									// fmt.Println("│ Scanning..")
									// fmt.Println("│ File : " + path) // Print matched file path
									// filehash, _ := calculateSHA256(path)
									// fmt.Println("│ Hash : " + filehash)
									// fmt.Println("│")

									// Read the file in path to a string ymlContent

									ymlContentBytes, err := os.ReadFile(path)
									if err != nil {
										LogError(fmt.Errorf("error reading file: %s", err))
										return nil
									}

									var tomlObj interface{}
									err = toml.Unmarshal(ymlContentBytes, &tomlObj)
									if err != nil {
										LogError(fmt.Errorf("error unmarshaling toml data: %v", err))
										return nil
									}

									ymlContent := string(ymlContentBytes)

									xvalid, _ := validateTOMLAndCUEContent(ymlContent, value.Toml_Structure)
									if xvalid {
										// fmt.Printf("│ The toml file is valid.\n")
										colorGreenBold.Println("│ ✓ ", value.ID, " ", path)
										stats.Clean++
										stats.Total++

									} else {

										// fmt.Printf("│ The toml file is not valid: %s\n", xerrMsg)
										colorRedBold.Println("│ ✗ ", value.ID, " ", path)
										stats.Dirty++
										stats.Total++
										warning = true

										envfound := FindMatchingString(cfgEnv, value.Environment, ",")
										if (envfound || strings.Contains(value.Environment, "all") || value.Environment == "") && value.Fatal {
											fatal = true
											stats.Fatal++
										}

									}

								}
								return nil
							})
							if err != nil {

								LogError(fmt.Errorf("error walking directory: %s", err))
							}

						} else {

							LogError(errors.New("TOML Required policy fields not set (File Pattern / TOML Structure)"))

						}
					}
				}(value) // Pass the loop variable to the goroutine
			}

			wg.Wait() // Wait for all goroutines to finish

		}
		if outputType == "full" || outputType == "sarif" {
			GenerateComplianceSarif(otCompliance)
		}

		fmt.Println("│")
		fmt.Println("│")
		fmt.Println("│")

		table := uitable.New()
		table.MaxColWidth = 254

		table.AddRow(colorBold.Render("├ Quick Stats "), "")
		table.AddRow(colorBold.Render("│"), "")
		table.AddRow(colorBold.Render("│ Total Policies Scanned"), ": "+colorBold.Render(stats.Total))
		table.AddRow(colorGreenBold.Render("│ Clean Policy Checks"), ": "+colorGreenBold.Render(stats.Clean))
		table.AddRow(colorYellowBold.Render("│ Irregularities Found"), ": "+colorYellowBold.Render(stats.Dirty))
		table.AddRow(colorRedBold.Render("│"), "")
		table.AddRow(colorRedBold.Render("│ Fatal Policy Breach"), ": "+colorRedBold.Render(stats.Fatal))

		fmt.Println(table)

		jsonstats, _jerr := json.Marshal(stats)
		if _jerr != nil {
			LogError(_jerr)
		}

		_jwerr := os.WriteFile("intercept.stats.json", jsonstats, 0644)
		if _jwerr != nil {
			LogError(_jwerr)
		}

		fmt.Println("│")
		fmt.Println("│")
		fmt.Println("│")
		fmt.Println("│")

		if fatal || stats.Total == 0 {

			colorRedBold.Println("│")
			colorRedBold.Println("├ ", rules.ExitCritical)
			colorRedBold.Println("│")
			if stats.Total == 0 {
				colorRedBold.Println("├──────── NO POLICIES WERE SCANNED ────────")
				colorRedBold.Println("│")
			}
			PrintClose()
			fmt.Println("")
			if tomlscanBreak != "false" {
				colorRedBold.Println("► break signal ")
				os.Exit(1)
			}
			os.Exit(0)

		}

		if warning {
			colorYellowBold.Println("│")
			colorYellowBold.Println("├ ", rules.ExitWarning)
			colorYellowBold.Println("│")

		} else {

			colorGreenBold.Println("│")
			colorGreenBold.Println("├ ", rules.ExitClean)
			colorGreenBold.Println("│")

		}

		fmt.Println("│")
		fmt.Println("│")
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		formattedTime = endTime.Format("2006-01-02 15:04:05")
		fmt.Println("│ Δ ", duration)
		fmt.Println("├ F ", formattedTime)
		PrintClose()

	},
}

func init() {

	tomlCmd.PersistentFlags().StringVarP(&tomlScanPath, "target", "t", ".", "scanning Target path")
	tomlCmd.PersistentFlags().StringVarP(&tomlscanTags, "tags", "i", "", "include only rules with the specified tag")
	tomlCmd.PersistentFlags().StringVarP(&tomlscanBreak, "break", "b", "true", "disable exit 1 for fatal rules")
	tomlCmd.PersistentFlags().StringVarP(&tomlTurboAPI, "silenturbo", "s", "false", "disable verbose output enabling turbo mode")

	rootCmd.AddCommand(tomlCmd)

}
