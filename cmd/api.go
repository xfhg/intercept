package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"os"
	"strings"

	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "INTERCEPT / API - Scan API endpoints from configured policy rules",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if cfgEnv == "" {
			cfgEnv = "先锋"
		}

		_ = os.Remove("intercept.output.json")
		_ = os.Remove("intercept.sarif.json")
		_ = os.Remove("intercept.scannedSHA256.json")

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
		//fmt.Println("│ Scan PATH\t: ", scanPath)
		fmt.Println("│ Scan ENV\t: ", cfgEnv)
		fmt.Println("│ Scan TAG\t: ", scanTags)

		// files, err := PathSHA256(scanPath)
		// if err != nil {
		// 	LogError(err)
		// }

		// jsonData, err := json.Marshal(struct {
		// 	Scanned []ScannedFile `json:"scanned"`
		// }{
		// 	Scanned: files,
		// })
		// if err != nil {
		// 	LogError(err)
		// 	return
		// }

		// err = os.WriteFile("intercept.scannedSHA256.json", jsonData, 0644)
		// if err != nil {
		// 	LogError(err)
		// }

		// fmt.Println("│ ")
		// fmt.Printf("│ Target Size : %d file(s)\n", len(files))

		// if len(files) < 20 {

		// 	fmt.Println("│ Target Filelist Checksum :")
		// 	fmt.Println("│ ")

		// 	for _, file := range files {
		// 		cleanPath := filepath.Clean(file.Path)
		// 		fmt.Printf("│ Path: %s :SHA256: %s\n", cleanPath, file.Sha256)

		// 	}
		// 	fmt.Println("│ ")
		// }
		// fmt.Println("│ ")
		fmt.Println(line)

		if auditNox {
			fmt.Println("│ Exceptions Disabled - All Policies Activated")
		}

		searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex"}, "")

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

		if scanTurbo == "false" && len(rules.Rules) < 50 {

			for _, value := range rules.Rules {

				tagfound := FindMatchingString(scanTags, value.Tags, ",")
				if tagfound || scanTags == "" {
					fmt.Println("│ ")
				} else {
					continue
				}

				searchPattern := []byte(strings.Join(value.Patterns, "\n") + "\n")
				_ = os.WriteFile(searchPatternFile, searchPattern, 0644)

				switch value.Type {

				case "api":

					gatheringData(value)
					processAPIType(value)

				default:

				}

			}

			_ = os.Remove(searchPatternFile)

			fmt.Println("│")
			fmt.Println("│")
			fmt.Println("│")

			colorBold.Println("├  Quick Stats : ")
			colorBold.Println("│ \t Total Policies Scanned :\t", stats.Total)
			colorGreenBold.Println("│ \t Clean Policy Checks :\t\t", stats.Clean)
			colorYellowBold.Println("│ \t Policy Irregularities :\t", stats.Dirty)
			colorRedBold.Println("│ \t Fatal Policy Breach :\t\t", stats.Fatal)

			jsonstats, _jerr := json.Marshal(stats)
			if _jerr != nil {
				LogError(_jerr)
			}

			_jwerr := os.WriteFile("stats.json", jsonstats, 0644)
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

		} else {

			// startTime = time.Now()
			// formattedTime = startTime.Format("2006-01-02 15:04:05")
			// fmt.Println("├ S ", formattedTime)
			// fmt.Print("│")

			// numCPU := runtime.NumCPU()

			// var wg sync.WaitGroup
			// wg.Add(len(rules.Rules))

			// sem := make(chan struct{}, numCPU*2)

			// rulesChan := make(chan Rule, len(rules.Rules))

			// for i := 0; i < numCPU; i++ {
			// 	go func(workerID int) {
			// 		for rule := range rulesChan {
			// 			sem <- struct{}{} // Acquire a token

			// 			tagfound := FindMatchingString(scanTags, rule.Tags, ",")
			// 			if tagfound || scanTags == "" {
			// 				worker(workerID, sem, &wg, rgembed, pwddir, scanPath, rule)
			// 			} else {
			// 				wg.Done() // Call wg.Done() if the worker skips the rule
			// 			}
			// 		}
			// 	}(i)
			// }

			// for _, policy := range rules.Rules {
			// 	rulesChan <- policy
			// }

			// close(rulesChan)

			// wg.Wait()
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

	rootCmd.AddCommand(apiCmd)

}
