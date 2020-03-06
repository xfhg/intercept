/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/ffmt.v1"
	"gopkg.in/yaml.v2"
)

var addCfgFile string
var defaultCfgFile string

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "INTERCEPT CONFIG - Add and merge config files to setup AUDIT",
	Long:  ``,

	Run: func(cmd *cobra.Command, args []string) {

		// if reset delete config file
		if configReset {
			if fileExists(defaultCfgFile) {
				_ = os.Remove(defaultCfgFile)
				fmt.Println("")
				fmt.Println("| INTERCEPT CONFIG")
				fmt.Println("|")
				fmt.Println("| Config clear")
				fmt.Println("")
			}
		}

		defaultCfgFile = "config.yaml"

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

				ffmt.Puts(master)

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

				ffmt.Puts(newfile)

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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
