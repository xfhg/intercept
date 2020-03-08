package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var addCfgFile string
var defaultCfgFile string

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "INTERCEPT CONFIG - Add and merge config files to setup AUDIT",
	Long:  ``,

	Run: func(cmd *cobra.Command, args []string) {

		defaultCfgFile = cfgFile

		if configReset {
			if fileExists(defaultCfgFile) {
				_ = os.Remove(defaultCfgFile)
				fmt.Println("|")
				fmt.Println("| Config clear")
			}
		}

		if fileExists(addCfgFile) {

			fmt.Println("|")
			fmt.Println("| Config file updated")
			fmt.Println("|")

			if fileExists(defaultCfgFile) {

				var master map[string]interface{}
				bs, err := ioutil.ReadFile(defaultCfgFile)
				if err != nil {
					panic(err)
				}
				if err := yaml.Unmarshal(bs, &master); err != nil {
					panic(err)
				}

				var override map[string]interface{}
				bs, err = ioutil.ReadFile(addCfgFile)
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

			} else {

				var newfile map[string]interface{}
				nf, err := ioutil.ReadFile(addCfgFile)
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

			}

		} else {

			fmt.Println("|")
			fmt.Println("| No updates detected")
			fmt.Println("|")

		}

	},
}

func init() {

	configCmd.PersistentFlags().BoolP("reset", "r", false, "Reset config file")
	configCmd.PersistentFlags().StringVarP(&addCfgFile, "add", "a", "", "Add config file (yaml)")

	rootCmd.AddCommand(configCmd)

}
