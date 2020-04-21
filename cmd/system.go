package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "INTERCEPT / SYSTEM - Setup, Check and Update system tools to run AUDIT",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		if systemSetup {

			fmt.Println("|")
			fmt.Println("| System Setup")
			fmt.Println("|")

			core := ""
			coreDst := GetExecutablePath()

			switch runtime.GOOS {
			case "windows":
				core = "https://github.com/xfhg/intercept/releases/latest/download/i-ripgrep-win.zip"
			case "darwin":
				core = "https://github.com/xfhg/intercept/releases/latest/download/i-ripgrep-macos.zip"
			case "linux":
				core = "https://github.com/xfhg/intercept/releases/latest/download/i-ripgrep-linux.zip"
			default:
				colorRedBold.Println("| OS not supported")
				PrintClose()
				os.Exit(1)
			}

			cfg := yacspin.Config{
				Frequency:       100 * time.Millisecond,
				CharSet:         yacspin.CharSets[51],
				Suffix:          " Downloading Core",
				SuffixAutoColon: true,
				Message:         core,
				StopCharacter:   "| ✓",
				StopColors:      []string{"fgGreen"},
			}

			spinner, err := yacspin.New(cfg)
			if err != nil {
				panic(err)
			}
			spinner.Start()

			//spinner.Message("data data")

			client := &getter.Client{
				Ctx:  context.Background(),
				Dst:  coreDst,
				Dir:  true,
				Src:  core,
				Mode: getter.ClientModeDir,
			}

			if err := client.Get(); err != nil {
				fmt.Fprintf(os.Stderr, "| Error getting path %s : %v", client.Src, err)
				PrintClose()
				os.Exit(1)
			}

			spinner.Message("Creating .ignore file")

			d := []string{"search_regex", "config.yaml"}
			err = WriteLinesOnFile(d, filepath.Join(coreDst, ".ignore"))
			err = WriteLinesOnFile(d, filepath.Join(GetWd(), ".ignore"))

			spinner.Stop()

		}

		if systemVersion {

			// FEATURE FLAG OFF

			latestVersion := ""
			exxecutablePath := GetExecutablePath()

			fmt.Println(exxecutablePath)

			workingPath := GetWd()

			fmt.Println(workingPath)

			fmt.Println("|")
			fmt.Println("| Latest version : ", latestVersion)
			fmt.Println("| ")

		}

		if !systemSetup && !systemVersion {

			fmt.Println("|")
			fmt.Println("| To setup/update the core system add --setup to this command")
			fmt.Println("| To check for intercept updates add --version to this command")
			fmt.Println("|")

		}

		PrintClose()

	},
}

func init() {

	systemCmd.PersistentFlags().BoolP("setup", "s", false, "Setup core tools")
	systemCmd.PersistentFlags().BoolP("version", "v", false, "validate system Version")

	rootCmd.AddCommand(systemCmd)

}
