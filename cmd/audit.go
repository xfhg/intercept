package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/gookit/color.v1"
)

var (
	scanPath string
	fatal    = false
	warning  = false
)

// AllRules : internal struct of the rules yaml file
type allRules struct {
	Exceptions []struct {
		Id   int    `yaml:"id"`
		Auth string `yaml:"auth"`
	} `yaml:"Exceptions"`
	Banner        string `yaml:"banner"`
	Exit_Critical string `yaml:"exit_critical"`
	Exit_Warning  string `yaml:"exit_warning"`
	Exit_Clean    string `yaml:"exit_clean"`
	Rules         []struct {
		Id          int      `yaml:"id"`
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Solution    string   `yaml:"solution"`
		Error       string   `yaml:"error"`
		Type        string   `yaml:"type"`
		Environment string   `yaml:"environment"`
		Fatal       bool     `yaml:"fatal"`
		Patterns    []string `yaml:"patterns"`
	} `yaml:"Rules"`
	Rules_Deactivated []int `yaml:"rules_deactivated"`
}

var (
	rules *allRules

	colorRedBold    = color.New(color.Red, color.OpBold)
	colorGreenBold  = color.New(color.Green, color.OpBold)
	colorYellowBold = color.New(color.Yellow, color.OpBold)
)

func getRuleStruct() *allRules {

	err := viper.Unmarshal(&rules)
	if err != nil {

		colorRedBold.Println("| Unable to decode config struct : ", err)

	}
	return rules

}

// auditCmd represents the audit command
var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "INTERCEPT / AUDIT - Scan a target path against configured policy rules and exceptions",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		rules = getRuleStruct()

		rgbin := "rg"

		path, err := exec.LookPath("rg")

		if err != nil || path == "" {

			switch runtime.GOOS {
			case "windows":
				rgbin = "rg/rg.exe"
			case "darwin":
				rgbin = "rg/rgm"
			case "linux":
				rgbin = "rg/rgl"
			default:

				colorRedBold.Println("| OS not supported")
				os.Exit(1)

			}
		}

		pwddir, _ := os.Getwd()

		searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex"}, "")

		fmt.Println("| ")
		fmt.Println("| ")
		fmt.Println(rules.Banner)
		fmt.Println("| ")

		for index, value := range rules.Rules {

			searchPattern := []byte(strings.Join(value.Patterns, "\n") + "\n")
			_ = ioutil.WriteFile(searchPatternFile, searchPattern, 0644)

			switch value.Type {

			case "scan":

				fmt.Println("| ")
				fmt.Println("| ------------------------------------------------------------ ", index)
				fmt.Println("| Rule #", value.Id)
				fmt.Println("| Rule name : ", value.Name)
				fmt.Println("| Rule description : ", value.Description)
				fmt.Println("| ")

				exception := sort.SearchInts(rules.Rules_Deactivated, value.Id)

				if exception < len(rules.Rules_Deactivated) && rules.Rules_Deactivated[exception] == value.Id {

					colorRedBold.Println("|")
					colorRedBold.Println("| THIS RULE CHECK IS DEACTIVATED BY AN EXCEPTION REQUEST ")
					colorRedBold.Println("|")

				} else {

					codePatternScan := []string{"--pcre2", "-p", "-i", "-C2", "-U", "-f", searchPatternFile, scanPath}
					xcmd := exec.Command(rgbin, codePatternScan...)

					xcmd.Stdout = os.Stdout
					xcmd.Stderr = os.Stderr

					errr := xcmd.Run()

					if errr != nil {
						if xcmd.ProcessState.ExitCode() == 2 {
							colorRedBold.Println("| Error")
							log.Fatal(errr)
						} else {
							colorGreenBold.Println("| Clean")
							fmt.Println("| ")
						}
					} else {

						if value.Environment == cfgEnv || value.Environment == "" {
							if value.Fatal {

								colorRedBold.Println("|")
								colorRedBold.Println("| FATAL : ")
								colorRedBold.Println(value.Error)
								colorRedBold.Println("|")
								fatal = true

							}
						} else {

							if value.Environment != cfgEnv && value.Environment != "" {

								colorRedBold.Println("|")
								colorRedBold.Println("|")
								colorRedBold.Println(value.Error)
								colorRedBold.Println("|")
								warning = true

							}
						}

						colorRedBold.Println("|")
						colorRedBold.Println("| Rule : ", value.Name)
						colorRedBold.Println("| Target Environment : ", value.Environment)
						colorRedBold.Println("| Suggested Solution : ", value.Solution)
						colorRedBold.Println("|")
						fmt.Println("| ")

					}
				}

			case "collect":

				fmt.Println("| ")
				fmt.Println("| ------------------------------------------------------------")
				fmt.Println("| Collection : ", value.Name)
				fmt.Println("| Description : ", value.Description)
				fmt.Println("| ")

				codePatternCollect := []string{"--pcre2", "--no-heading", "-i", "-o", "-U", "-f", searchPatternFile, scanPath}
				xcmd := exec.Command(rgbin, codePatternCollect...)

				xcmd.Stdout = os.Stdout
				xcmd.Stderr = os.Stderr
				err := xcmd.Run()

				if err != nil {
					if xcmd.ProcessState.ExitCode() == 2 {
						colorRedBold.Println("| Error")
						log.Fatal(err)
					} else {
						colorRedBold.Println("| Clean")
						fmt.Println("| ")
					}
				} else {
					fmt.Println("| ")
				}

			default:

			}

		}

		_ = os.Remove(searchPatternFile)

		fmt.Println("|")
		fmt.Println("|")
		fmt.Println("|")

		if fatal {

			colorRedBold.Println("| ", rules.Exit_Critical)
			fmt.Println("|")
			fmt.Println("| INTERCEPT")
			fmt.Println("")
			colorRedBold.Println("- break signal - ")
			os.Exit(1)
		}

		if warning {

			colorYellowBold.Println("| ", rules.Exit_Warning)

		} else {

			colorGreenBold.Println("| ", rules.Exit_Clean)

		}

		fmt.Println("|")
		fmt.Println("| INTERCEPT")
		fmt.Println("")
	},
}

func init() {

	auditCmd.PersistentFlags().StringVarP(&scanPath, "target", "t", ".", "scanning Target path")
	auditCmd.PersistentFlags().BoolP("report", "d", false, "debrief json Report output file (auditdebrief.json)")

	rootCmd.AddCommand(auditCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// auditCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// auditCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
