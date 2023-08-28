package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	yamlTurboAPI  string
	yamlScanPath  string
	yamlscanTags  string
	yamlscanBreak string
)

var yamlCmd = &cobra.Command{
	Use:   "yml",
	Short: "INTERCEPT / YAML - Scan YAML structure from configured policy rules",
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
		fmt.Println("│ Scan TAG\t: ", yamlscanTags)
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

		/*

			1. detect if target is file or folder
			2. collect all files defined by the pile pattern regex
			3. detect if they are yml
			4. split the filestructure
			5. loop it inside the file and find the structure value
			6. validate the found value against the ASSURE pattern
			7. print the compliant file name , file line , and sha256 of the compliant value
			8. list uncompliant files

		*/

		// step one detect if folder

		// loop rules
		if yamlTurboAPI == "false" && len(rules.Rules) < 50 {

			for _, value := range rules.Rules {

				tagfound := FindMatchingString(yamlscanTags, value.Tags, ",")
				if tagfound || yamlscanTags == "" {
					fmt.Println("│ ")
				} else {
					continue
				}

				switch value.Type {

				case "yml":

					if value.Yml_Filepattern != "" && value.Yml_Structure != "" {

						fmt.Println("│ ")
						fmt.Println(line)
						fmt.Println("│ ")
						fmt.Println("├ YML Rule #", value.ID)
						fmt.Println("│ Rule name : ", value.Name)
						fmt.Println("│ Rule description : ", value.Description)
						fmt.Println("│ Impacted Env : ", value.Environment)
						fmt.Println("│ Confidence : ", value.Confidence)
						fmt.Println("│ Tags : ", value.Tags)
						fmt.Println("│ File pattern : ", value.Yml_Filepattern)
						fmt.Println("│ Yml pattern : ", value.Yml_Structure)
						fmt.Println("│ ")

						// Compile regex pattern
						regex, err := regexp.Compile(value.Yml_Filepattern)
						if err != nil {
							LogError(fmt.Errorf("Error compiling regex: %s", err))
							return
						}

						fileInfo, err := os.Stat(yamlScanPath)
						if err != nil {
							LogError(fmt.Errorf("Error accessing path: %s", err))
						}

						// Check if the path is a directory
						if fileInfo.IsDir() {
							// fmt.Printf("%s is a directory\n", yamlScanPath)
						} else {
							LogError(fmt.Errorf("%s is not a directory", yamlScanPath))
						}

						// Function to match files
						err = filepath.Walk(yamlScanPath, func(path string, info os.FileInfo, err error) error {
							if err != nil {
								return err
							}
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
									LogError(fmt.Errorf("Error reading file: %s", err))
									return nil
								}

								var yamlObj interface{}
								err = yaml.Unmarshal(ymlContentBytes, &yamlObj)
								if err != nil {
									LogError(fmt.Errorf("Error unmarshaling YAML data: %v", err))
									return nil
								}

								ymlContent := string(ymlContentBytes)

								xvalid, xerrMsg := validateYAMLAndCUEContent(ymlContent, value.Yml_Structure)
								if xvalid {
									// fmt.Printf("│ The YAML file is valid.\n")
									colorGreenBold.Println("│ ")
									colorGreenBold.Println("│ Compliant")
									colorGreenBold.Println("│ ")
									colorGreenBold.Println("├── ─")
									fmt.Println("│")

								} else {
									// fmt.Printf("│ The YAML file is not valid: %s\n", xerrMsg)
									colorRedBold.Println("│")
									colorRedBold.Println("│ NON COMPLIANT : ")
									colorRedBold.Println("│ ", value.Error)
									colorRedBold.Println("│ ", xerrMsg)
									colorRedBold.Println("│")
									colorRedBold.Println("├── ─")
									fmt.Println("│")
								}

							}
							return nil
						})
						if err != nil {

							LogError(fmt.Errorf("Error walking directory: %s", err))
						}

					} else {

						LogError(errors.New("YML Required policy fields not set (File Pattern / Yml Structure)"))

					}

				default:

				}

			}

		} else {
		}

		PrintClose()

	},
}

func init() {

	yamlCmd.PersistentFlags().StringVarP(&yamlScanPath, "target", "t", ".", "scanning Target path")
	yamlCmd.PersistentFlags().StringVarP(&yamlscanTags, "tags", "i", "", "include only rules with the specified tag")
	yamlCmd.PersistentFlags().StringVarP(&yamlscanBreak, "break", "b", "true", "disable exit 1 for fatal rules")
	yamlCmd.PersistentFlags().StringVarP(&yamlTurboAPI, "silenturbo", "s", "false", "disable verbose output enabling turbo mode")

	rootCmd.AddCommand(yamlCmd)

}
