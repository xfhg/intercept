package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "INTERCEPT / SYSTEM - Setup, Check and Update system tools to run AUDIT",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		// FEATURE FLAG OFF

		if systemSetup {
			fmt.Println("|")
			fmt.Println("| System Setup")
		}

	},
}

func init() {

	systemCmd.PersistentFlags().BoolP("setup", "s", false, "Setup essential tools")

	rootCmd.AddCommand(systemCmd)

}
