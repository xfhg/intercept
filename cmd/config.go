package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var addCfgFile string
var hashCfgFile string
var defaultCfgFile string

const (
	HTTPPrefix     = "http://"
	HTTPSPrefix    = "https://"
	ConfigFilename = "config.yaml"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "INTERCEPT / CONFIG - Add and merge policy and config files",
	Long:  ``,

	Run: func(cmd *cobra.Command, args []string) {

		var err error
		var fromURL bool
		var fromFile bool
		var downloadedCfgFile io.ReadCloser

		defaultCfgFile = cfgFile
		fromURL = false

		if configReset {
			if FileExists(defaultCfgFile) {
				_ = os.Remove(defaultCfgFile)

			}
			fmt.Println("│")
			colorGreenBold.Println("│ Config clear")

		}

		if strings.HasPrefix(addCfgFile, HTTPPrefix) || strings.HasPrefix(addCfgFile, HTTPSPrefix) {

			downloadedCfgFile, err = ReaderFromURL(addCfgFile)

			if err != nil {
				colorRedBold.Println("│ Error")
				log.Fatal(err)
			}
			defer downloadedCfgFile.Close()

			if err != nil {
				colorRedBold.Println("│ Error")
				log.Fatal(err)
			}
			fromURL = true

			fmt.Println("│")
			fmt.Println("│ Config downloaded")

		}

		if FileExists(addCfgFile) {
			fromFile = true

			fmt.Println("│")
			fmt.Println("│ Config read from file")

		}

		if FileExists(defaultCfgFile) && (fromFile || fromURL) {
			// merge
			var master map[string]interface{}
			bs, err := os.ReadFile(defaultCfgFile)
			if err != nil {
				LogError(err)
			}
			if err := yaml.Unmarshal(bs, &master); err != nil {
				LogError(err)
			}

			var override map[string]interface{}
			if fromURL {
				bs, err = io.ReadAll(downloadedCfgFile)
			} else {
				bs, err = os.ReadFile(addCfgFile)
			}
			if err != nil {
				LogError(err)
			}

			HexDigestConfig := sha256hash(bs)

			if hashCfgFile != "" {

				fmt.Println("│")
				fmt.Println("│ SHA256 Valid checksum :\t", hashCfgFile)
				fmt.Println("│ SHA256 Config checksum :\t", HexDigestConfig)

				if HexDigestConfig != hashCfgFile {
					colorRedBold.Println("│")
					colorRedBold.Println("│ Error")
					colorRedBold.Println("│")
					log.Fatal("Aborting : MD5 checksum does not match")
				} else {
					fmt.Println("│")
					colorGreenBold.Println("│ MD5 Config Match")
					fmt.Println("│")
				}
			}

			if err := yaml.Unmarshal(bs, &override); err != nil {
				LogError(err)
			}

			for k, v := range override {
				if strings.Contains(k, "Rules") {
					fmt.Println("│ Protected from rewriting [Rules]")
					fmt.Println("│ This component is declared once, cannot be merged")
				} else {
					master[k] = v
				}

			}

			bs, err = yaml.Marshal(master)
			if err != nil {
				LogError(err)
			}
			if err := os.WriteFile(ConfigFilename, bs, 0644); err != nil {
				LogError(err)
			}

			fmt.Println("│")
			fmt.Println("│ Config Updated")
			fmt.Println("│")
			fmt.Println("└")

		} else if fromFile || fromURL {
			// new
			var newfile map[string]interface{}
			var nf []byte
			var err error

			if fromURL {
				nf, err = io.ReadAll(downloadedCfgFile)
			} else {
				nf, err = os.ReadFile(addCfgFile)
			}

			if err != nil {
				LogError(err)
			}

			HexDigestConfig := sha256hash(nf)

			if hashCfgFile != "" {

				fmt.Println("│")
				fmt.Println("│ SHA256 Expected checksum :\t", hashCfgFile)
				fmt.Println("│ SHA256 Config checksum :\t", HexDigestConfig)

				if HexDigestConfig != hashCfgFile {
					colorRedBold.Println("│")
					colorRedBold.Println("│ Error")
					colorRedBold.Println("│")
					log.Fatal("Aborting : MD5 checksum does not match")
				} else {
					fmt.Println("│")
					colorGreenBold.Println("│ MD5 Config Hash Match")
					fmt.Println("│")
				}
			}

			if err := yaml.Unmarshal(nf, &newfile); err != nil {
				LogError(err)
			}
			nf, err = yaml.Marshal(newfile)
			if err != nil {
				LogError(err)
			}
			if err := os.WriteFile("config.yaml", nf, 0644); err != nil {
				LogError(err)
			}
			fmt.Println("│")
			colorGreenBold.Println("│ New Config created")
			fmt.Println("│")
			fmt.Println("└")

		} else {
			fmt.Println("│")
			fmt.Println("│ No updates detected")
			fmt.Println("│")
			fmt.Println("└")
		}

	},
}

func init() {

	configCmd.PersistentFlags().BoolP("reset", "r", false, "Reset config file")
	configCmd.PersistentFlags().StringVarP(&addCfgFile, "add", "a", "", "Add config file (yaml)")
	configCmd.PersistentFlags().StringVarP(&hashCfgFile, "hash", "k", "", "Config file SHA256 Hash")

	rootCmd.AddCommand(configCmd)

}
