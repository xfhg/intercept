package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// main command flags
var cfgFile string
var cfgEnv string
var outputType string

// subcommand flags
var (
	configReset  bool
	configb64    bool
	auditNox     bool
	systemSetup  bool
	systemUpdate bool
	buildVersion string
)

var rootCmd = &cobra.Command{
	Use:   "intercept",
	Short: "INTERCEPT / Policy as Code / Audit & Compliance",
	Long:  ``,
}

// Execute is the global command
func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "global Config file (config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&cfgEnv, "environment", "e", "", "global Environment id")
	rootCmd.PersistentFlags().StringVarP(&outputType, "output", "o", "full", "output type full/json/sarif/console")

}

func initConfig() {

	viper.SetConfigType("yaml")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home := GetHomeDir()
		viper.AddConfigPath(home)
		viper.SetConfigFile("config.yaml")
	}

	viper.AutomaticEnv()

	PrintStart()

	if err := viper.ReadInConfig(); err == nil {

		fmt.Println("├ Loaded PaC as ", viper.ConfigFileUsed())

	}

	configReset = configCmdisReset()
	configb64 = configCmdisB64()
	auditNox = auditCmdisNoExceptions()
	systemSetup = systemCmdisSetup()
	systemUpdate = systemCmdisUpdate()

}

func systemCmdisSetup() bool {

	setup, _ := systemCmd.PersistentFlags().GetBool("setup")
	return setup
}

func systemCmdisUpdate() bool {

	update, _ := systemCmd.PersistentFlags().GetBool("update")
	return update
}

func configCmdisReset() bool {

	b64, _ := configCmd.PersistentFlags().GetBool("reset")
	return b64
}

func configCmdisB64() bool {

	reset, _ := configCmd.PersistentFlags().GetBool("b64")
	return reset
}

func auditCmdisNoExceptions() bool {

	nox, _ := auditCmd.PersistentFlags().GetBool("no-exceptions")
	return nox
}
