package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func processAssureType(value Rule) {

	if cfgEnv == "" {
		cfgEnv = "先锋"
	}

	rules := loadUpRules()

	pwddir := GetWd()

	rgembed, _ := prepareEmbeddedExecutable()

	searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex"}, "")

	fmt.Println("│ ")
	fmt.Println(line)
	fmt.Println("│ ")
	fmt.Println("├ ASSURE Rule #", value.ID)
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
		xcmd := exec.Command(rgembed, codePatternScan...)
		xcmd.Stdout = os.Stdout
		xcmd.Stderr = os.Stderr
		errr := xcmd.Run()

		if errr != nil {
			if xcmd.ProcessState.ExitCode() == 2 {
				LogError(errr)
			} else {

				envfound := FindMatchingString(cfgEnv, value.Environment, ",")
				if (envfound || strings.Contains(value.Environment, "all") || value.Environment == "") && value.Fatal {

					colorRedBold.Println("│")
					colorRedBold.Println("│ NON COMPLIANT : ")
					colorRedBold.Println("│ ", value.Error)
					colorRedBold.Println("│")
					fatal = true
					stats.Fatal++
				} else {

					colorRedBold.Println("│")
					colorRedBold.Println("│ NOT FOUND")
					colorRedBold.Println("│ ", value.Error)
					colorRedBold.Println("│")
					warning = true

				}
				colorRedBold.Println("│")
				colorRedBold.Println("│ ASSURE Rule : ", value.Name)
				colorRedBold.Println("│ Target Environment : ", value.Environment)
				colorRedBold.Println("│ Suggested Solution : ", value.Solution)
				colorRedBold.Println("│")
				fmt.Println("│ ")
				stats.Total++
				stats.Dirty++

			}
		} else {

			colorGreenBold.Println("│ ")
			colorGreenBold.Println("│ Compliant")
			stats.Clean++
			stats.Total++
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

		codePatternScanJSON := []string{"--pcre2", "--no-heading", "-o", "-p", "-i", "-U", "--json", "-f", searchPatternFile, scanPath}
		xcmdJSON := exec.Command(rgembed, codePatternScanJSON...)
		xcmdJSON.Stdout = jsonoutfile
		xcmdJSON.Stderr = os.Stderr
		errrJSON := xcmdJSON.Run()

		if errrJSON != nil {
			if xcmdJSON.ProcessState.ExitCode() == 2 {
				LogError(errrJSON)
			} else {

				ProcessOutput(strings.Join([]string{strconv.Itoa(value.ID), ".json"}, ""), strconv.Itoa(value.ID), value.Name, value.Description, value.Error, value.Solution, value.Fatal)
				GenerateSarif()
				colorRedBold.Println("│ ")
			}
		} else {
			colorGreenBold.Println("│")
			os.Remove(jsonOutputFile)

		}

	}

}
