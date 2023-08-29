package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var (
	tomlTurboAPI  string
	tomlScanPath  string
	tomlscanTags  string
	tomlscanBreak string
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

								ymlContent := string(ymlContentBytes)

								xvalid, xerrMsg := validateTOMLAndCUEContent(ymlContent, value.Toml_Structure)
								if xvalid {
									// fmt.Printf("│ The toml file is valid.\n")
									colorGreenBold.Println("│ ")
									colorGreenBold.Println("│ Compliant")
									colorGreenBold.Println("│ ")
									colorGreenBold.Println("├── ─")
									fmt.Println("│")

								} else {
									// fmt.Printf("│ The toml file is not valid: %s\n", xerrMsg)
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

							LogError(fmt.Errorf("error walking directory: %s", err))
						}

					} else {

						LogError(errors.New("TOML Required policy fields not set (File Pattern / TOML Structure)"))

					}

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

									} else {
										// fmt.Printf("│ The toml file is not valid: %s\n", xerrMsg)
										colorRedBold.Println("│ ✗ ", value.ID, " ", path)

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
