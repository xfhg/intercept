package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
)

var (
	assureTurboAPI  string
	assureScanPath  string
	assurescanTags  string
	assurescanBreak string
	aCompliance     InterceptComplianceOutput
	aRule           InterceptCompliance
)

var assureCmd = &cobra.Command{
	Use:   "assure",
	Short: "INTERCEPT / ASSURE - Enforce configured policy rules",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if cfgEnv == "" {
			cfgEnv = "先锋"
		}

		cleanupFiles()

		rules = loadUpRules()

		pwddir := GetWd()

		stats = allStats{0, 0, 0, 0}

		startTime := time.Now()

		formattedTime := startTime.Format("2006-01-02 15:04:05")

		fmt.Println("├ S ", formattedTime)
		fmt.Println("│")

		line := "├────────────────────────────────────────────────────────────"

		fmt.Println("│ ")
		fmt.Println("│ Scan PATH\t: ", assureScanPath)
		fmt.Println("│ Scan ENV\t: ", cfgEnv)
		fmt.Println("│ Scan TAG\t: ", assurescanTags)
		fmt.Println("│ ")
		// fmt.Println(line)

		files, err := PathSHA256(assureScanPath)
		if err != nil {
			LogError(err)
		}

		jsonData, err := json.Marshal(struct {
			Scanned []ScannedFile `json:"scanned"`
		}{
			Scanned: files,
		})
		if err != nil {
			LogError(err)
			return
		}

		err = os.WriteFile("intercept.scannedSHA256.json", jsonData, 0644)
		if err != nil {
			LogError(err)
		}

		fmt.Println("│ ")
		fmt.Printf("│ Target Size : %d file(s)\n", len(files))

		if len(files) < 20 {

			shatable := uitable.New()
			shatable.MaxColWidth = 254

			shatable.AddRow(colorBold.Render("├ Target Filelist"), "")
			shatable.AddRow(colorBold.Render("│ Path"), colorBold.Render("SHA256 Checksum"))

			for _, file := range files {
				cleanPath := filepath.Clean(file.Path)
				shatable.AddRow("│ "+cleanPath, file.Sha256)
			}

			fmt.Println(shatable)
			fmt.Println("│ ")
		}
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

		if assureTurboAPI == "false" && len(rules.Rules) < 50 {

			for _, value := range rules.Rules {

				tagfound := FindMatchingString(scanTags, value.Tags, ",")
				if tagfound || scanTags == "" {
					fmt.Println("│ ")
				} else {
					continue
				}

				searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex_", value.ID}, "")

				searchPattern := []byte(strings.Join(value.Patterns, "\n") + "\n")
				_ = os.WriteFile(searchPatternFile, searchPattern, 0644)

				switch value.Type {

				case "assure":

					processAssureType(value)

				default:

				}

				_ = os.Remove(searchPatternFile)

			}

			GenerateSarif("assure")
			GenerateComplianceSarif(aCompliance)
			GenerateAssureSARIF()

		} else {

			// var wg sync.WaitGroup

			fmt.Println("├─ Turbo Scanning not yet available ─ ─")

			// for _, value := range rules.Rules {

			// 	wg.Add(1) // Increment the WaitGroup counter

			// 	go func(value Rule) { // Launch a goroutine with the loop value
			// 		defer wg.Done() // Decrement the counter when the goroutine completes

			// 		tagfound := FindMatchingString(assurescanTags, value.Tags, ",")
			// 		if tagfound || assurescanTags == "" {
			// 			fmt.Println("│ ")
			// 		} else {
			// 			return
			// 		}

			// 		switch value.Type {

			// 		case "assure":

			// 		}
			// 	}(value)
			// }
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

		assurestats, _jerr := json.Marshal(stats)
		if _jerr != nil {
			LogError(_jerr)
		}

		_jwerr := os.WriteFile("intercept.stats.json", assurestats, 0644)
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
			if assurescanBreak != "false" {
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

	assureCmd.PersistentFlags().StringVarP(&assureScanPath, "target", "t", ".", "scanning Target path")
	assureCmd.PersistentFlags().StringVarP(&assurescanTags, "tags", "i", "", "include only rules with the specified tag")
	assureCmd.PersistentFlags().StringVarP(&assurescanBreak, "break", "b", "true", "disable exit 1 for fatal rules")
	assureCmd.PersistentFlags().StringVarP(&assureTurboAPI, "silenturbo", "s", "false", "disable verbose output enabling turbo mode")

	rootCmd.AddCommand(assureCmd)

}
