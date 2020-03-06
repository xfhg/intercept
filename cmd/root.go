package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// main command flags
var cfgFile string
var cfgEnv string

// subcommand flags
var (
	configReset bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "intercept",
	Short: "INTERCEPT / Policy as Code Static Analysis Auditing",
	Long:  ``,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.yaml", "global Config file (config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&cfgEnv, "environment", "e", "", "global Environment id")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.SetConfigType("yaml")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".intercept" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigFile("config.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	fmt.Println("")
	fmt.Println("| INTERCEPT")
	fmt.Println("|")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {

		fmt.Println("| Loading policy file :", viper.ConfigFileUsed())

	}

	// Collect all subcommand flags here
	configReset = configCmdisReset()

}

func configCmdisReset() bool {

	reset, _ := configCmd.PersistentFlags().GetBool("reset")
	return reset
}
