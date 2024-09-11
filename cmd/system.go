package cmd

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var testEmbeddedCmd = &cobra.Command{
	Use:   "sys",
	Short: "Test intercept embedded core binaries",
	Long:  `This command extracts and runs the embedded rg and goss binaries to verify they are working correctly.`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Debug().Msg("Core binaries:")
		log.Debug().Msgf("rg path: %s", rgPath)
		log.Debug().Msgf("goss path: %s", gossPath)

		files, err := os.ReadDir(filepath.Dir(rgPath))
		if err != nil {
			log.Debug().Msgf("Error reading directory %s: %v", filepath.Dir(rgPath), err)
			return
		}
		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				log.Debug().Msgf("Error getting info for file %s: %v", file.Name(), err)
				continue
			}
			log.Debug().Msgf("File: %s, Permissions: %s", info.Name(), info.Mode().Perm())
		}

		files, err = os.ReadDir(filepath.Dir(gossPath))
		if err != nil {
			log.Debug().Msgf("Error reading directory %s: %v", filepath.Dir(rgPath), err)
			return
		}

		for _, file := range files {
			info, err := file.Info()
			if err != nil {
				log.Debug().Msgf("Error getting info for file %s: %v", file.Name(), err)
				continue
			}
			log.Debug().Msgf("File: %s, Permissions: %s", info.Name(), info.Mode().Perm())
		}

		// Test rg
		log.Debug().Msg("Testing rg:")
		rgCmd := exec.Command(rgPath, "--version")
		rgCmd.Stdout = os.Stdout
		rgCmd.Stderr = os.Stderr
		if err := rgCmd.Run(); err != nil {
			log.Debug().Msgf("Error running rg: %v", err)
		}

		// Test goss
		log.Debug().Msg("Testing goss:")
		gossCmd := exec.Command(gossPath, "--version")
		gossCmd.Stdout = os.Stdout
		gossCmd.Stderr = os.Stderr
		if err := gossCmd.Run(); err != nil {
			log.Debug().Msgf("Error running goss: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(testEmbeddedCmd)
}
