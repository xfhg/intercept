package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update called")
		confirmAndSelfUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

}

func confirmAndSelfUpdate() {
	latest, found, err := selfupdate.DetectLatest("xfhg/intercept")
	if err != nil {
		log.Println("Error occurred while detecting version:", err)
		return
	}

	v := semver.MustParse(buildVersion)
	if !found || latest.Version.LTE(v) {
		log.Println("Current version is the latest")
		return
	}

	fmt.Print("Do you want to update to", latest.Version, "? (y/n): ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		log.Println("Invalid input")
		return
	}
	if input == "n\n" {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		log.Println("Could not locate executable path")
		return
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		log.Println("Error occurred while updating binary:", err)
		return
	}
	log.Println("Successfully updated to version", latest.Version)

}
