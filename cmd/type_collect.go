package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func processCollectType(value Rule) {

	if cfgEnv == "" {
		cfgEnv = "先锋"
	}

	pwddir := GetWd()

	rgembed, _ := prepareEmbeddedExecutable()

	searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex_", value.ID}, "")

	fmt.Println("│ ")
	fmt.Println(line)
	fmt.Println("│ ")
	fmt.Println("├ COLLECT Rule #", value.ID)
	fmt.Println("├ Name : ", value.Name)
	fmt.Println("│ Description : ", value.Description)
	fmt.Println("│ Tags : ", value.Tags)
	fmt.Println("│ ")

	codePatternCollect := []string{"--pcre2", "--no-heading", "-i", "-o", "-U", "-f", searchPatternFile, scanPath}
	xcmd := exec.Command(rgembed, codePatternCollect...)
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

	jsonOutputFile := strings.Join([]string{pwddir, "/", value.ID, ".json"}, "")
	jsonoutfile, erroutjson := os.Create(jsonOutputFile)
	if erroutjson != nil {
		LogError(erroutjson)
	}
	defer jsonoutfile.Close()
	writer := bufio.NewWriter(jsonoutfile)
	defer writer.Flush()

	codePatternScanJSON := []string{"--pcre2", "--no-heading", "-i", "-o", "-U", "--json", "-f", searchPatternFile, scanPath}
	xcmdJSON := exec.Command(rgembed, codePatternScanJSON...)
	xcmdJSON.Stdout = jsonoutfile
	xcmdJSON.Stderr = os.Stderr
	errrJSON := xcmdJSON.Run()

	if errrJSON != nil {
		if xcmdJSON.ProcessState.ExitCode() == 2 {
			LogError(errrJSON)
		} else {
			colorGreenBold.Println("│")
			os.Remove(jsonOutputFile)
		}
	} else {
		ProcessOutput(strings.Join([]string{value.ID, ".json"}, ""), value.ID, value.Type, value.Name, value.Description, "", "", false)
		colorRedBold.Println("│ ")
	}
}
