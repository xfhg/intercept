package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"

	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	scanPath  string
	scanTags  string
	scanBreak string
	fatal     = false
	warning   = false
)

type allStats struct {
	Total int `json:"Total"`
	Clean int `json:"Clean"`
	Dirty int `json:"Dirty"`
	Fatal int `json:"Fatal"`
}

type allRules struct {
	Banner string `yaml:"banner"`

	ExceptionMessage string `yaml:"exceptionmessage"`
	ExitCritical     string `yaml:"exitcritical"`
	ExitWarning      string `yaml:"exitwarning"`
	ExitClean        string `yaml:"exitclean"`

	Rules []struct {
		ID          int      `yaml:"id"`
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Solution    string   `yaml:"solution"`
		Error       string   `yaml:"error"`
		Type        string   `yaml:"type"`
		Environment string   `yaml:"environment"`
		Enforcement bool     `yaml:"enforcement"`
		Fatal       bool     `yaml:"fatal"`
		Tags        string   `yaml:"tags,omitempty"`
		Impact      string   `yaml:"impact,omitempty"`
		Confidence  string   `yaml:"confidence,omitempty"`
		Patterns    []string `yaml:"patterns"`
	} `yaml:"Rules"`

	Exceptions []int `yaml:"exceptions"`
}

var (
	stats           allStats
	rules           *allRules
	colorRedBold    = color.New(color.Red, color.OpBold)
	colorGreenBold  = color.New(color.Green, color.OpBold)
	colorYellowBold = color.New(color.Yellow, color.OpBold)
	colorBlueBold   = color.New(color.Blue, color.OpBold)
	colorBold       = color.New(color.OpBold)
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

		_ = os.Remove("intercept.output.json")
		_ = os.Remove("intercept.sarif.json")

		rules = loadUpRules()

		stats = allStats{0, 0, 0, 0}

		pwddir := GetWd()

		rgbin := CoreExists()

		fmt.Println("│ PWD : ", pwddir)
		fmt.Println("│ RGP : ", rgbin)
		fmt.Println("│ ")
		fmt.Println("│ Scan PATH : ", scanPath)
		fmt.Println("│ Scan ENV : ", cfgEnv)
		fmt.Println("│ Scan TAG : ", scanTags)

		if auditNox {
			fmt.Println("│ Exceptions Disabled - All Policies Activated")
		}

		line := "├────────────────────────────────────────────────────────────"

		searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex"}, "")

		fmt.Println("│ ")
		fmt.Println("│ ")
		if rules.Banner != "" {
			fmt.Println(rules.Banner)
		}
		fmt.Println("│ ")
		if len(rules.Rules) < 1 {
			fmt.Println("│ No Policy rules detected")
			fmt.Println("│ Run the command - intercept config - to setup policies")
			PrintClose()
			os.Exit(0)
		}

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

			case "scan":

				fmt.Println("│ ")
				fmt.Println(line)
				fmt.Println("│ ")
				fmt.Println("├ Rule #", value.ID)
				fmt.Println("│ Rule name : ", value.Name)
				fmt.Println("│ Rule description : ", value.Description)
				fmt.Println("│ Impacted Env : ", value.Environment)
				fmt.Println("│ Confidence : ", value.Confidence)
				fmt.Println("│ Tags : ", value.Tags)
				fmt.Println("│ ")

				exception := ContainsInt(rules.Exceptions, value.ID)

				if exception && !auditNox && !value.Enforcement {

					colorRedBold.Println("│")
					colorRedBold.Println("│ ", rules.ExceptionMessage)
					colorRedBold.Println("│")

				} else {

					codePatternScan := []string{"--pcre2", "-p", "-o", "-A0", "-B0", "-C0", "-i", "-U", "-f", searchPatternFile, scanPath}
					xcmd := exec.Command(rgbin, codePatternScan...)
					xcmd.Stdout = os.Stdout
					xcmd.Stderr = os.Stderr
					errr := xcmd.Run()

					if errr != nil {
						if xcmd.ProcessState.ExitCode() == 2 {
							LogError(errr)
						} else {
							colorGreenBold.Println("│ Clean")
							stats.Clean++
							stats.Total++
							fmt.Println("│ ")
						}
					} else {

						envfound := FindMatchingString(cfgEnv, value.Environment, ",")
						if (envfound || strings.Contains(value.Environment, "all") || value.Environment == "") && value.Fatal {

							colorRedBold.Println("│")
							colorRedBold.Println("│ FATAL : ")
							colorRedBold.Println("│ ", value.Error)
							colorRedBold.Println("│")
							fatal = true
							stats.Fatal++
						} else {

							colorRedBold.Println("│")
							colorRedBold.Println("│")
							colorRedBold.Println("│ ", value.Error)
							colorRedBold.Println("│")
							warning = true

						}
						colorRedBold.Println("│")
						colorRedBold.Println("│ Rule : ", value.Name)
						colorRedBold.Println("│ Target Environment : ", value.Environment)
						colorRedBold.Println("│ Suggested Solution : ", value.Solution)
						colorRedBold.Println("│")
						fmt.Println("│ ")
						stats.Total++
						stats.Dirty++
					}

					jsonOutputFile := strings.Join([]string{pwddir, "/", strconv.Itoa(value.ID), ".json"}, "")
					jsonoutfile, erroutjson := os.Create(jsonOutputFile)
					if erroutjson != nil {
						LogError(erroutjson)
					}
					defer jsonoutfile.Close()
					writer := bufio.NewWriter(jsonoutfile)
					defer writer.Flush()

					codePatternScanJSON := []string{"--pcre2", "--no-heading", "-o", "-p", "-i", "-U", "--json", "-f", searchPatternFile, scanPath}
					xcmdJSON := exec.Command(rgbin, codePatternScanJSON...)
					xcmdJSON.Stdout = jsonoutfile
					xcmdJSON.Stderr = os.Stderr
					errrJSON := xcmdJSON.Run()

					if errrJSON != nil {
						if xcmdJSON.ProcessState.ExitCode() == 2 {
							LogError(errrJSON)
						} else {
							colorRedBold.Println("│ ")
						}
					} else {
						ProcessOutput(strings.Join([]string{strconv.Itoa(value.ID), ".json"}, ""), strconv.Itoa(value.ID), value.Name, value.Description, value.Error, value.Solution, value.Fatal)
						GenerateSarif()
						colorBlueBold.Println("│ ")

					}

				}

			case "collect":

				fmt.Println("│ ")
				fmt.Println(line)
				fmt.Println("│ ")
				fmt.Println("├ Collection : ", value.Name)
				fmt.Println("│ Description : ", value.Description)
				fmt.Println("│ Tags : ", value.Tags)
				fmt.Println("│ ")

				codePatternCollect := []string{"--pcre2", "--no-heading", "-i", "-o", "-U", "-f", searchPatternFile, scanPath}
				xcmd := exec.Command(rgbin, codePatternCollect...)
				xcmd.Stdout = os.Stdout
				xcmd.Stderr = os.Stderr
				err := xcmd.Run()

				if err != nil {
					if xcmd.ProcessState.ExitCode() == 2 {
						LogError(err)
					} else {
						colorGreenBold.Println("│ Clean")
						fmt.Println("│ ")
					}
				} else {
					fmt.Println("│ ")
				}

				jsonOutputFile := strings.Join([]string{pwddir, "/", strconv.Itoa(value.ID), ".json"}, "")
				jsonoutfile, erroutjson := os.Create(jsonOutputFile)
				if erroutjson != nil {
					LogError(erroutjson)
				}
				defer jsonoutfile.Close()
				writer := bufio.NewWriter(jsonoutfile)
				defer writer.Flush()

				codePatternScanJSON := []string{"--pcre2", "--no-heading", "-i", "-o", "-U", "--json", "-f", searchPatternFile, scanPath}
				xcmdJSON := exec.Command(rgbin, codePatternScanJSON...)
				xcmdJSON.Stdout = jsonoutfile
				xcmdJSON.Stderr = os.Stderr
				errrJSON := xcmdJSON.Run()

				if errrJSON != nil {
					if xcmdJSON.ProcessState.ExitCode() == 2 {
						LogError(errrJSON)
					} else {
						colorBlueBold.Println("│")
					}
				} else {
					ProcessOutput(strings.Join([]string{strconv.Itoa(value.ID), ".json"}, ""), strconv.Itoa(value.ID), value.Name, value.Description, "", "", false)
					colorRedBold.Println("│ ")
				}

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
			if scanBreak != "true" {
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

		PrintClose()
	},
}

func init() {

	auditCmd.PersistentFlags().StringVarP(&scanPath, "target", "t", ".", "scanning Target path")
	auditCmd.PersistentFlags().BoolP("no-exceptions", "x", false, "disables the option to deactivate rules by eXception")
	auditCmd.PersistentFlags().StringVarP(&scanTags, "tags", "i", "", "include only rules with the specified tag")
	auditCmd.PersistentFlags().StringVarP(&scanBreak, "no-break", "b", "false", "disable exit 1 for fatal rules")

	rootCmd.AddCommand(auditCmd)

}
