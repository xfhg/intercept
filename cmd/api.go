package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"os"
	"strings"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var (
	scanTurboAPI  string
	apiCompliance InterceptComplianceOutput
	apiRule       InterceptCompliance
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "INTERCEPT / API - Scan API endpoints from configured policy rules",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if cfgEnv == "" {
			cfgEnv = "先锋"
		}

		cleanupFiles()

		rules = loadUpRules()

		stats = allStats{0, 0, 0, 0}

		pwddir := GetWd()

		rgembed, _ := prepareEmbeddedExecutable()

		startTime := time.Now()
		formattedTime := startTime.Format("2006-01-02 15:04:05")
		fmt.Println("├ S ", formattedTime)
		fmt.Println("│")

		line := "├────────────────────────────────────────────────────────────"

		fmt.Println("│ PWD : ", pwddir)
		fmt.Println("│ RGP : ", rgembed)
		fmt.Println("│ ")

		fmt.Println("│ Scan ENV\t: ", cfgEnv)
		fmt.Println("│ Scan TAG\t: ", scanTags)
		fmt.Println("│ ")
		fmt.Println(line)

		if auditNox {
			fmt.Println("│ Exceptions Disabled - All Policies Activated")
		}

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

		if scanTurboAPI == "false" && len(rules.Rules) < 50 {

			for _, value := range rules.Rules {

				tagfound := FindMatchingString(scanTags, value.Tags, ",")
				if tagfound || scanTags == "" {
					fmt.Println("│ ")
				} else {
					continue
				}

				switch value.Type {

				case "api":

					searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex_", strconv.Itoa(value.ID)}, "")

					searchPattern := []byte(strings.Join(value.Patterns, "\n") + "\n")
					_ = os.WriteFile(searchPatternFile, searchPattern, 0600)

					gatheringData(value, false)
					processAPIType(value, false)

				default:

				}

			}

			GenerateSarif("api")
			GenerateComplianceSarif(apiCompliance)
			GenerateApiSARIF()

		} else {

			values := make(chan Rule, len(rules.Rules))
			var wg sync.WaitGroup

			for _, value := range rules.Rules {

				switch value.Type {

				case "api":

					searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex_", strconv.Itoa(value.ID)}, "")
					searchPattern := []byte(strings.Join(value.Patterns, "\n") + "\n")
					_ = os.WriteFile(searchPatternFile, searchPattern, 0600)

					tagfound := FindMatchingString(scanTags, value.Tags, ",")
					if tagfound || scanTags == "" {
						wg.Add(1)
						go func(value Rule) {
							defer wg.Done()
							gatheringData(value, true)
							values <- value
						}(value)
					}

				default:

				}
			}

			wg.Wait()
			close(values)

			for i := 0; i < len(rules.Rules); i++ {

				wg.Add(1)
				go func() {
					defer wg.Done()
					value := <-values
					switch value.Type {

					case "api":
						processAPIType(value, true)
					default:

					}

				}()

			}

			wg.Wait()

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

		_jwerr := os.WriteFile("intercept.stats.json", jsonstats, 0600)
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
			if scanBreak != "false" {
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

		endTime := time.Now()
		duration := endTime.Sub(startTime)
		formattedTime = endTime.Format("2006-01-02 15:04:05")
		fmt.Println("│ Δ ", duration)
		fmt.Println("├ F ", formattedTime)
		PrintClose()

	},
}

func init() {

	apiCmd.PersistentFlags().StringVarP(&scanTags, "tags", "i", "", "include only rules with the specified tag")
	apiCmd.PersistentFlags().StringVarP(&scanBreak, "break", "b", "true", "disable exit 1 for fatal rules")
	apiCmd.PersistentFlags().StringVarP(&scanTurboAPI, "silenturbo", "s", "false", "disable verbose output enabling turbo mode")

	rootCmd.AddCommand(apiCmd)

}
