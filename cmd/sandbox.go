package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/spf13/cobra"
	"github.com/theckman/yacspin"
)

var targetFile string

var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "INTERCEPT / SANDBOX - experimental features",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("│")
		fmt.Println("├ Sandbox")
		fmt.Println("│")

		err := DownloadJSONFile(targetFile, "downloaded.file")
		if err != nil {
			LogError(err)
		}

		PrintClose()

	},
}

func init() {

	sandboxCmd.PersistentFlags().BoolP("activate", "l", false, "Activate Sandbox Features")

	sandboxCmd.PersistentFlags().StringVarP(&targetFile, "dl", "d", "", "Download target file")

	rootCmd.AddCommand(sandboxCmd)

}

func DownloadJSONFile(src, dst string) error {
	// Send a HEAD request to get the Content-Type without downloading the body
	resp, err := http.Head(src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the Content-Type is JSON
	// if resp.Header.Get("Content-Type") != "application/json" {
	// 	return errors.New("source is not a JSON file")
	// }

	// Configure the getter client
	client := &getter.Client{
		Src:  src,
		Dst:  dst,
		Mode: getter.ClientModeFile,
	}

	// Configure the yacspin spinner
	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[3],
		Suffix:          " Downloading",
		SuffixAutoColon: true,
		StopCharacter:   "│ ✓",
		StopColors:      []string{"fgGreen"},
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		return err
	}

	// Start the spinner
	err = spinner.Start()
	if err != nil {
		return err
	}

	// Download the file
	err = client.Get()
	if err != nil {
		spinner.StopFail()
		return err
	}

	// Stop the spinner
	spinner.Stop()

	// Read the downloaded file
	data, err := ioutil.ReadFile(dst)
	if err != nil {
		return err
	}

	// Compute the SHA-256 hash of the file
	hash := sha256hash(data)

	fmt.Println("│")
	fmt.Println("├ Hash: ", hash)
	fmt.Println("│")

	return nil
}
