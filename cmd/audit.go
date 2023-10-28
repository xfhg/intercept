package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/gosuri/uitable"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scanPath  string
	scanTags  string
	scanBreak string
	scanTurbo string
	fatal     = false
	warning   = false
)

var line = "├────────────────────────────────────────────────────────────"

type allStats struct {
	Total int `json:"Total"`
	Clean int `json:"Clean"`
	Dirty int `json:"Dirty"`
	Fatal int `json:"Fatal"`
}

type ScannedFile struct {
	Path   string `json:"path"`
	Sha256 string `json:"sha256"`
}

type Rule struct {
	ID               int      `yaml:"id"`
	Name             string   `yaml:"name"`
	Description      string   `yaml:"description"`
	Solution         string   `yaml:"solution"`
	Error            string   `yaml:"error"`
	Type             string   `yaml:"type"`
	Environment      string   `yaml:"environment"`
	Enforcement      bool     `yaml:"enforcement"`
	Fatal            bool     `yaml:"fatal"`
	Tags             string   `yaml:"tags,omitempty"`
	Impact           string   `yaml:"impact,omitempty"`
	Confidence       string   `yaml:"confidence,omitempty"`
	Api_Endpoint     string   `yaml:"api_endpoint,omitempty"`
	Api_Request      string   `yaml:"api_request,omitempty"`
	Api_Insecure     bool     `yaml:"api_insecure"`
	Api_Body         string   `yaml:"api_body,omitempty"`
	Api_Auth         string   `yaml:"api_auth,omitempty"`
	Api_Auth_Basic   *string  `yaml:"api_auth_basic,omitempty"`
	Api_Auth_Token   *string  `yaml:"api_auth_token,omitempty"`
	Api_Trace        bool     `yaml:"api_trace,omitempty"`
	Filepattern      string   `yaml:"filepattern,omitempty"`
	Yml_Filepattern  string   `yaml:"yml_filepattern,omitempty"`
	Yml_Structure    string   `yaml:"yml_structure,omitempty"`
	Toml_Filepattern string   `yaml:"toml_filepattern,omitempty"`
	Toml_Structure   string   `yaml:"toml_structure,omitempty"`
	Json_Filepattern string   `yaml:"json_filepattern,omitempty"`
	Json_Structure   string   `yaml:"json_structure,omitempty"`
	Patterns         []string `yaml:"patterns"`
}

type allRules struct {
	Banner           string `yaml:"banner"`
	Rules            []Rule `yaml:"Rules"`
	ExitCritical     string `yaml:"exitcritical"`
	ExitWarning      string `yaml:"exitwarning"`
	ExitClean        string `yaml:"exitclean"`
	Exceptions       []int  `yaml:"exceptions"`
	ExceptionMessage string `yaml:"exceptionmessage"`
}

var (
	stats           allStats
	rules           *allRules
	colorRedBold    = color.New(color.Red, color.OpBold)
	colorGreenBold  = color.New(color.Green, color.OpBold)
	colorYellowBold = color.New(color.Yellow, color.OpBold)
	colorBlueBold   = color.New(color.Blue, color.OpBold)
	colorBold       = color.New(color.OpBold)
	colorYellow     = color.New(color.Yellow)
)

func loadUpRules() *allRules {

	err := viper.Unmarshal(&rules)
	if err != nil {
		colorRedBold.Println("│ Unable to decode config struct : ", err)
	}
	return rules

}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "INTERCEPT / AUDIT - Scan a target path against configured policy rules",
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
		fmt.Println("│ Scan PATH\t: ", scanPath)
		fmt.Println("│ Scan ENV\t: ", cfgEnv)
		fmt.Println("│ Scan TAG\t: ", scanTags)

		files, err := PathSHA256(scanPath)
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

		if scanTurbo == "false" && len(rules.Rules) < 50 {

			for _, value := range rules.Rules {

				tagfound := FindMatchingString(scanTags, value.Tags, ",")
				if tagfound || scanTags == "" {
					fmt.Println("│ ")
				} else {
					continue
				}

				searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex_", strconv.Itoa(value.ID)}, "")

				searchPattern := []byte(strings.Join(value.Patterns, "\n") + "\n")
				_ = os.WriteFile(searchPatternFile, searchPattern, 0644)

				switch value.Type {

				case "api":

					// moved to its own command - will compile here later
					// gatheringData(value, false)
					// processAPIType(value, false)

				case "assure":

					// moved to its own command - will compile here later
					// processAssureType(value)

				case "scan":

					processScanType(value)

				case "collect":

					processCollectType(value)

				default:

				}

				_ = os.Remove(searchPatternFile)

			}

			GenerateSarif("audit")

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

			startTime = time.Now()
			formattedTime := startTime.Format("2006-01-02 15:04:05")
			fmt.Println("├ S ", formattedTime)
			fmt.Print("│")

			numCPU := runtime.NumCPU()

			var wg sync.WaitGroup
			wg.Add(len(rules.Rules))

			sem := make(chan struct{}, numCPU*2)

			rulesChan := make(chan Rule, len(rules.Rules))

			for i := 0; i < numCPU; i++ {
				go func(workerID int) {
					for rule := range rulesChan {
						sem <- struct{}{} // Acquire a token

						tagfound := FindMatchingString(scanTags, rule.Tags, ",")
						if tagfound || scanTags == "" {
							worker(workerID, sem, &wg, rgembed, pwddir, scanPath, rule)
						} else {
							wg.Done() // Call wg.Done() if the worker skips the rule
						}
					}
				}(i)
			}

			for _, policy := range rules.Rules {
				rulesChan <- policy
			}

			close(rulesChan)

			wg.Wait()
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

	auditCmd.PersistentFlags().StringVarP(&scanPath, "target", "t", ".", "scanning Target path")
	auditCmd.PersistentFlags().BoolP("no-exceptions", "x", false, "disables the option to deactivate rules by eXception")
	auditCmd.PersistentFlags().StringVarP(&scanTags, "tags", "i", "", "include only rules with the specified tag")
	auditCmd.PersistentFlags().StringVarP(&scanBreak, "break", "b", "true", "disable exit 1 for fatal rules")
	auditCmd.PersistentFlags().StringVarP(&scanTurbo, "silenturbo", "s", "false", "disable verbose output enabling turbo mode")

	rootCmd.AddCommand(auditCmd)

}
