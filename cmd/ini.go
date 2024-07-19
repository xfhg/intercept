package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ini/ini"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var (
	iniTurboAPI  string
	iniScanPath  string
	iniscanTags  string
	iniscanBreak string
	oiCompliance InterceptComplianceOutput
	oiRule       InterceptCompliance
)

var iniCmd = &cobra.Command{
	Use:   "ini",
	Short: "INTERCEPT / INI - Scan INI structure from configured policy rules",
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
		fmt.Println("│ Scan TAG\t: ", iniscanTags)
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

		if iniTurboAPI == "false" && len(rules.Rules) < 50 {

			for _, value := range rules.Rules {

				tagfound := FindMatchingString(iniscanTags, value.Tags, ",")
				if tagfound || iniscanTags == "" {
					fmt.Println("│ ")
				} else {
					continue
				}

				switch value.Type {

				case "ini":

					if value.Ini_Filepattern != "" && value.Ini_Structure != "" {

						fmt.Println("│ ")
						fmt.Println(line)
						fmt.Println("│ ")
						fmt.Println("├ Ini Rule #", value.ID)
						fmt.Println("│ Rule name : ", value.Name)
						fmt.Println("│ Rule description : ", value.Description)
						fmt.Println("│ Impacted Env : ", value.Environment)
						fmt.Println("│ Confidence : ", value.Confidence)
						fmt.Println("│ Tags : ", value.Tags)
						fmt.Println("│ File pattern : ", value.Ini_Filepattern)
						fmt.Println("│ INI pattern : ", value.Ini_Structure)
						fmt.Println("│ ")

						oiRule = InterceptCompliance{}
						oiRule.RuleDescription = value.Description
						oiRule.RuleError = value.Error
						oiRule.RuleFatal = value.Fatal
						oiRule.RuleID = strconv.Itoa(value.ID)
						oiRule.RuleName = value.Name
						oiRule.RuleSolution = value.Solution
						oiRule.RuleType = value.Type

						// Compile regex pattern
						regex, err := regexp.Compile(value.Ini_Filepattern)
						if err != nil {
							LogError(fmt.Errorf("error compiling regex: %s", err))
							return
						}

						fileInfo, err := os.Stat(iniScanPath)
						if err != nil {
							LogError(fmt.Errorf("error accessing path: %s", err))
						}

						// Check if the path is a directory
						if fileInfo.IsDir() {
							// fmt.Printf("%s is a directory\n", iniScanPath)
						} else {
							LogError(fmt.Errorf("%s is not a directory", iniScanPath))
						}

						// Function to match files
						err = filepath.Walk(iniScanPath, func(path string, info os.FileInfo, err error) error {
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

								//var iniObj interface{}
								_, err = ini.Load(ymlContentBytes)
								if err != nil {
									LogError(fmt.Errorf("error unmarshaling INI data: %v", err))
									return nil
								}

								oFinding = InterceptComplianceFinding{
									FileName: path,
									FileHash: filehash,
									ParentID: value.ID,
								}

								ymlContent := string(ymlContentBytes)

								xvalid, xerrMsg := validateINIAndCUEContent(ymlContent, value.Ini_Structure)
								if xvalid {
									// fmt.Printf("│ The INI file is valid.\n")
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

									// fmt.Printf("│ The ini file is not valid: %s\n", xerrMsg)
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
								oiRule.RuleFindings = append(oiRule.RuleFindings, oFinding)

							}
							return nil
						})
						if err != nil {

							LogError(fmt.Errorf("error walking directory: %s", err))
						}

					} else {

						LogError(errors.New("INI Required policy fields not set (File Pattern / INI Structure)"))

					}

					oiCompliance = append(oiCompliance, oiRule)

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

					tagfound := FindMatchingString(iniscanTags, value.Tags, ",")
					if tagfound || iniscanTags == "" {
						fmt.Println("│ ")
					} else {
						return
					}

					switch value.Type {

					case "ini":

						if value.Ini_Filepattern != "" && value.Ini_Structure != "" {

							regex, err := regexp.Compile(value.Ini_Filepattern)
							if err != nil {
								LogError(fmt.Errorf("error compiling regex: %s", err))
								return
							}

							fileInfo, err := os.Stat(iniScanPath)
							if err != nil {
								LogError(fmt.Errorf("error accessing path: %s", err))
							}

							// Check if the path is a directory
							if fileInfo.IsDir() {
								// fmt.Printf("%s is a directory\n", iniScanPath)
							} else {
								LogError(fmt.Errorf("%s is not a directory", iniScanPath))
							}

							// Function to match files
							err = filepath.Walk(iniScanPath, func(path string, info os.FileInfo, err error) error {
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

									//var iniObj interface{}
									_, err = ini.Load(ymlContentBytes)
									if err != nil {
										LogError(fmt.Errorf("error unmarshaling ini data: %v", err))
										return nil
									}

									ymlContent := string(ymlContentBytes)

									xvalid, _ := validateINIAndCUEContent(ymlContent, value.Ini_Structure)
									if xvalid {
										// fmt.Printf("│ The Ini file is valid.\n")
										colorGreenBold.Println("│ ✓ ", value.ID, " ", path)
										stats.Clean++
										stats.Total++

									} else {

										// fmt.Printf("│ The Ini file is not valid: %s\n", xerrMsg)
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

							LogError(errors.New("INI Required policy fields not set (File Pattern / INI Structure)"))

						}
					}
				}(value) // Pass the loop variable to the goroutine
			}

			wg.Wait() // Wait for all goroutines to finish

		}

		GenerateComplianceSarif(oiCompliance)

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

		if fatal {

			colorRedBold.Println("│")
			colorRedBold.Println("├ ", rules.ExitCritical)
			colorRedBold.Println("│")
			PrintClose()
			fmt.Println("")
			if iniscanBreak != "false" {
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

	iniCmd.PersistentFlags().StringVarP(&iniScanPath, "target", "t", ".", "scanning Target path")
	iniCmd.PersistentFlags().StringVarP(&iniscanTags, "tags", "i", "", "include only rules with the specified tag")
	iniCmd.PersistentFlags().StringVarP(&iniscanBreak, "break", "b", "true", "disable exit 1 for fatal rules")
	iniCmd.PersistentFlags().StringVarP(&iniTurboAPI, "silenturbo", "s", "false", "disable verbose output enabling turbo mode")

	rootCmd.AddCommand(iniCmd)

}
