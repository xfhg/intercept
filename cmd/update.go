package cmd

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		confirmAndSelfUpdate()
		PrintClose()

	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

}

func confirmAndSelfUpdate() {

	var current string

	fmt.Println("|")

	latest, found, err := selfupdate.DetectLatest("xfhg/intercept")
	if err != nil {
		fmt.Println("| Error occurred while detecting version:", err)
		return
	}
	if buildVersion != "" {
		current = buildVersion[1 : len(buildVersion)+1]
	} else {
		current = "0.0.1"
	}
	v := semver.MustParse(current)
	if !found || latest.Version.LTE(v) {
		fmt.Println("| Current version is the latest")
		return
	}

	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[51],
		Suffix:          " Downloading Update",
		SuffixAutoColon: true,
		Message:         latest.Version.String(),
		StopCharacter:   "| ✓",
		StopColors:      []string{"fgGreen"},
	}

	fmt.Print("| Do you want to update to v", latest.Version, " ? (y/n): ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		fmt.Println("| Invalid input")
		return
	}
	if input == "n\n" {
		return
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println("|")
	spinner.Start()

	exe, err := os.Executable()
	if err != nil {
		spinner.StopColors("fgRed")
		spinner.StopCharacter("| x")
		spinner.Message("Could not locate executable path")
		spinner.Stop()
		// fmt.Println("|")
		// fmt.Println("| Could not locate executable path")
		return
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		spinner.StopColors("fgRed")
		spinner.StopCharacter("| x")
		spinner.Message("Error occurred while updating binary")
		spinner.Stop()
		fmt.Println("|")
		fmt.Println("| Error occurred while updating binary:", err)
		return
	}

	spinner.Stop()

	fmt.Println("|")
	fmt.Println("| Successfully updated to version", latest.Version)

}
