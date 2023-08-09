package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "INTERCEPT / SYSTEM - Setup + Update intercept and its core system tools to run AUDIT",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if !systemUpdate && !systemSetup {

			fmt.Println("│")
			fmt.Println("│ Use : ")
			fmt.Println("│")
			fmt.Println("│ --update to download the latest intercept binary ")

		}

		if systemUpdate {

			selfUpdate()

		}

		PrintClose()

	},
}

func init() {

	systemCmd.PersistentFlags().BoolP("update", "u", false, "Update to the lastest version")

	rootCmd.AddCommand(systemCmd)

}

func selfUpdate() {

	var current string

	fmt.Println("│")
	fmt.Println("├ Binary Update")
	fmt.Println("│")

	latest, found, err := selfupdate.DetectLatest("xfhg/intercept")
	if err != nil {
		fmt.Println("│ Error occurred while detecting version:", err)
		return
	}
	if buildVersion != "" {
		current = buildVersion[1:]
	} else {
		current = "0.0.1"
	}
	v := semver.MustParse(current)
	if !found || latest.Version.LTE(v) {
		fmt.Println("│ Current version is the latest")
		return
	}

	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[3],
		Suffix:          " Downloading Update",
		SuffixAutoColon: true,
		Message:         latest.Version.String(),
		StopCharacter:   "│ ✓",
		StopColors:      []string{"fgGreen"},
	}

	fmt.Println("│ Updating to", latest.Version)

	spinner, err := yacspin.New(cfg)
	if err != nil {
		LogError(err)
	}
	fmt.Println("│")
	spinner.Start()

	exe, err := os.Executable()
	if err != nil {
		spinner.StopColors("fgRed")
		spinner.StopCharacter("│ x")
		spinner.Message("Could not locate executable path")
		spinner.Stop()
		return
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		spinner.StopColors("fgRed")
		spinner.StopCharacter("│ x")
		spinner.Message("Error occurred while updating binary")
		spinner.Stop()
		fmt.Println("│")
		fmt.Println("│ Error occurred while updating binary:", err)
		return
	}

	spinner.Stop()

	fmt.Println("│")
	fmt.Println("│ Successfully updated to version", latest.Version)

}
