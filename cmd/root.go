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

// subcommand flags
var (
	configReset  bool
	auditNox     bool
	systemSetup  bool
	updateAuto   bool
	buildVersion string
)

var rootCmd = &cobra.Command{
	Use:   "intercept",
	Short: "INTERCEPT / Policy as Code Static Analysis Auditor",
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

		fmt.Println("| Policy file :", viper.ConfigFileUsed())

	}

	configReset = configCmdisReset()
	auditNox = auditCmdisNoExceptions()
	systemSetup = systemCmdisSetup()
	updateAuto = updateCmdAuto()

}

func systemCmdisSetup() bool {

	setup, _ := systemCmd.PersistentFlags().GetBool("setup")
	return setup
}

func configCmdisReset() bool {

	reset, _ := configCmd.PersistentFlags().GetBool("reset")
	return reset
}

func auditCmdisNoExceptions() bool {

	nox, _ := auditCmd.PersistentFlags().GetBool("no-exceptions")
	return nox
}

func updateCmdAuto() bool {
	autoUpdate, _ := updateCmd.PersistentFlags().GetBool("auto")
	return autoUpdate
}
