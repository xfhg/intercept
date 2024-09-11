package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the build info of intercept",
	Long:  `Print the build version number of intercept along with its signature`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msgf("Intercept build version %s", buildVersion)
		log.Info().Msgf("Intercept signature [%s]", buildSignature)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
