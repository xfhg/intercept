package cmd

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

func worker(id int, sem chan struct{}, wg *sync.WaitGroup, rgbin string, pwddir string, scanPath string, policy Rule) {
	defer wg.Done()
	sem <- struct{}{}
	ripTurbo(rgbin, pwddir, scanPath, policy)
	<-sem
}

func ripTurbo(rgbin string, pwddir string, scanPath string, policy Rule) {

	jsonOutputFile := strings.Join([]string{pwddir, "/", strconv.Itoa(policy.ID), ".json"}, "")
	jsonoutfile, erroutjson := os.Create(jsonOutputFile)
	if erroutjson != nil {
		LogError(erroutjson)
	}
	defer jsonoutfile.Close()
	writer := bufio.NewWriter(jsonoutfile)
	defer writer.Flush()

	searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex_", strconv.Itoa(policy.ID)}, "")
	searchPattern := []byte(strings.Join(policy.Patterns, "\n") + "\n")
	_ = os.WriteFile(searchPatternFile, searchPattern, 0644)

	codePatternScanJSON := []string{"--pcre2", "--no-heading", "-o", "-p", "-i", "-U", "--json", "-f", searchPatternFile, scanPath}

	xcmdJSON := exec.Command(rgbin, codePatternScanJSON...)
	xcmdJSON.Stdout = jsonoutfile
	xcmdJSON.Stderr = os.Stdin

	errrJSON := xcmdJSON.Run()

	os.Remove(searchPatternFile)

	if errrJSON != nil {

		if xcmdJSON.ProcessState.ExitCode() == 2 {
			LogError(errrJSON)
		} else {
			colorGreenBold.Print("│")
			os.Remove(jsonOutputFile)
		}

	} else {
		ProcessOutput(strings.Join([]string{strconv.Itoa(policy.ID), ".json"}, ""), strconv.Itoa(policy.ID), policy.Name, policy.Description, policy.Error, policy.Solution, policy.Fatal)
		GenerateSarif()
		colorRedBold.Print("│")
	}

}
