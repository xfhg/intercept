package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
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

type allRules struct {
	Banner           string `yaml:"banner"`
	ExceptionMessage string `yaml:"exceptionmessage"`
	ExitCritical     string `yaml:"exitcritical"`
	ExitWarning      string `yaml:"exitwarning"`
	ExitClean        string `yaml:"exitclean"`
	Rules            []struct {
		ID          int      `yaml:"id"`
		Name        string   `yaml:"name"`
		Description string   `yaml:"description"`
		Solution    string   `yaml:"solution"`
		Error       string   `yaml:"error"`
		Type        string   `yaml:"type"`
		Environment string   `yaml:"environment"`
		Enforcement bool     `yaml:"enforcement"`
		Fatal       bool     `yaml:"fatal"`
		Patterns    []string `yaml:"patterns"`
	} `yaml:"Rules"`
	RulesDeactivated []int `yaml:"rulesdeactivated"`
}

var (
	rules           *allRules
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

// ContainsInt
func containsInt(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

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

		fmt.Println("| Scan Path : ", scanPath)

		if auditNox {
			fmt.Println("| Exceptions Disabled ")
			fmt.Println("| All Policies Activated ")
		}

		pwddir, _ := os.Getwd()

		line := "------------------------------------------------------------"

		searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex"}, "")

		fmt.Println("| ")
		fmt.Println("| ")
		if rules.Banner != "" {
			fmt.Println(rules.Banner)
		}
		fmt.Println("| ")
		if len(rules.Rules) < 1 {
			fmt.Println("| No Policy rules detected")
		}

		for _, value := range rules.Rules {

			searchPattern := []byte(strings.Join(value.Patterns, "\n") + "\n")
			_ = ioutil.WriteFile(searchPatternFile, searchPattern, 0644)

			switch value.Type {

			case "scan":

				fmt.Println("| ")
				fmt.Println("|", line)
				fmt.Println("| ")
				fmt.Println("| Rule #", value.ID)
				fmt.Println("| Rule name : ", value.Name)
				fmt.Println("| Rule description : ", value.Description)
				fmt.Println("| ")

				exception := containsInt(rules.RulesDeactivated, value.ID)

				if exception && !auditNox && !value.Enforcement {

					colorRedBold.Println("|")
					colorRedBold.Println("| ", rules.ExceptionMessage)
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

						if strings.Contains(value.Environment, cfgEnv) ||
							strings.Contains(value.Environment, "all") ||
							value.Environment == "" {
							if value.Fatal {

								colorRedBold.Println("|")
								colorRedBold.Println("| FATAL : ")
								colorRedBold.Println("| ", value.Error)
								colorRedBold.Println("|")
								fatal = true

							}
						} else {

							if value.Environment != cfgEnv &&
								value.Environment != "" {

								colorRedBold.Println("|")
								colorRedBold.Println("|")
								colorRedBold.Println("| ", value.Error)
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
				fmt.Println("|", line)
				fmt.Println("| ")
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
						colorGreenBold.Println("| Clean")
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

			colorRedBold.Println("| ", rules.ExitCritical)
			fmt.Println("|")
			fmt.Println("| INTERCEPT")
			fmt.Println("")
			colorRedBold.Println("- break signal - ")
			os.Exit(1)
		}

		if warning {

			colorYellowBold.Println("| ", rules.ExitWarning)

		} else {

			colorGreenBold.Println("| ", rules.ExitClean)

		}

		fmt.Println("|")
		fmt.Println("| INTERCEPT")
		fmt.Println("")
	},
}

func init() {

	auditCmd.PersistentFlags().StringVarP(&scanPath, "target", "t", ".", "scanning Target path")
	auditCmd.PersistentFlags().BoolP("no-exceptions", "x", false, "disables the option to deactivate rules by exception")

	rootCmd.AddCommand(auditCmd)

}
