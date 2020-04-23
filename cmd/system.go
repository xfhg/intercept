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
	Short: "INTERCEPT / SYSTEM - Setup and Update core system tools to run AUDIT",
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
				core = "https://github.com/xfhg/intercept/releases/latest/download/i-ripgrep-x86_64-windows.zip"
			case "darwin":
				core = "https://github.com/xfhg/intercept/releases/latest/download/i-ripgrep-x86_64-darwin.zip"
			case "linux":
				core = "https://github.com/xfhg/intercept/releases/latest/download/i-ripgrep-x86_64-linux.zip"
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
				LogError(err)
			}
			spinner.Start()

			client := &getter.Client{
				Ctx:  context.Background(),
				Dst:  coreDst,
				Dir:  true,
				Src:  core,
				Mode: getter.ClientModeDir,
			}

			if err := client.Get(); err != nil {
				spinner.StopColors("fgRed")
				spinner.StopCharacter("| x")
				spinner.Message("Error getting path")
				spinner.Stop()
				fmt.Fprintf(os.Stderr, "| Error getting path %s : %v", client.Src, err)
				PrintClose()
				os.Exit(1)
			}

			spinner.Message("Creating .ignore file")

			d := []string{"search_regex", "config.yaml"}
			err = WriteLinesOnFile(d, filepath.Join(coreDst, ".ignore"))
			err = WriteLinesOnFile(d, filepath.Join(GetWd(), ".ignore"))

			spinner.Stop()

		} else {

			fmt.Println("|")
			fmt.Println("| To Setup/update the core system add --setup to this command")
			fmt.Println("| To Update intercept binary run $ intercept update")
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
