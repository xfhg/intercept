package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var addCfgFile string
var defaultCfgFile string

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "INTERCEPT / CONFIG - Add and merge config files to setup AUDIT",
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
			fmt.Println("|")
			fmt.Println("| Config clear")

		}

		if strings.HasPrefix(addCfgFile, "http://") || strings.HasPrefix(addCfgFile, "https://") {

			downloadedCfgFile, err = ReaderFromURL(addCfgFile)

			if err != nil {
				colorRedBold.Println("| Error")
				log.Fatal(err)
			}
			defer downloadedCfgFile.Close()

			if err != nil {
				colorRedBold.Println("| Error")
				log.Fatal(err)
			}
			fromURL = true

			fmt.Println("|")
			fmt.Println("| Config downloaded")

		}

		if FileExists(addCfgFile) {
			fromFile = true

			fmt.Println("|")
			fmt.Println("| Config read from file")

		}

		if FileExists(defaultCfgFile) && (fromFile || fromURL) {
			// merge
			var master map[string]interface{}
			bs, err := ioutil.ReadFile(defaultCfgFile)
			if err != nil {
				panic(err)
			}
			if err := yaml.Unmarshal(bs, &master); err != nil {
				panic(err)
			}

			var override map[string]interface{}
			if fromURL {
				bs, err = ioutil.ReadAll(downloadedCfgFile)
			} else {
				bs, err = ioutil.ReadFile(addCfgFile)
			}
			if err != nil {
				panic(err)
			}
			if err := yaml.Unmarshal(bs, &override); err != nil {
				panic(err)
			}

			for k, v := range override {
				master[k] = v
			}

			bs, err = yaml.Marshal(master)
			if err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile("config.yaml", bs, 0644); err != nil {
				panic(err)
			}

			fmt.Println("|")
			fmt.Println("| Config Updated")
			fmt.Println("└")

		} else if fromFile || fromURL {
			// new
			var newfile map[string]interface{}
			var nf []byte
			var err error

			if fromURL {
				nf, err = ioutil.ReadAll(downloadedCfgFile)
			} else {
				nf, err = ioutil.ReadFile(addCfgFile)
			}

			if err != nil {
				panic(err)
			}
			if err := yaml.Unmarshal(nf, &newfile); err != nil {
				panic(err)
			}
			nf, err = yaml.Marshal(newfile)
			if err != nil {
				panic(err)
			}
			if err := ioutil.WriteFile("config.yaml", nf, 0644); err != nil {
				panic(err)
			}
			fmt.Println("|")
			fmt.Println("| New Config created")
			fmt.Println("└")

		} else {
			fmt.Println("|")
			fmt.Println("| No updates detected")
			fmt.Println("└")
		}

	},
}

func init() {

	configCmd.PersistentFlags().BoolP("reset", "r", false, "Reset config file")
	configCmd.PersistentFlags().StringVarP(&addCfgFile, "add", "a", "", "Add config file (yaml)")

	rootCmd.AddCommand(configCmd)

}
