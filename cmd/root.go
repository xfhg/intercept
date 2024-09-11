package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack/v3"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
	"github.com/spf13/cobra"
)

var (
	intercept_run_id string
	verbosity        int
	outputDir        string
	experimentalMode bool
	silentMode       bool
	nologMode        bool

	buildVersion   string
	buildSignature string

	rootCmd = &cobra.Command{
		Use:   "intercept",
		Short: "DevSecOps toolkit",
		Long:  `Code Compliance`,
	}
	log zerolog.Logger
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error executing root command")
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "increase verbosity level")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output-dir", "o", "", "directory to write output files")
	rootCmd.PersistentFlags().BoolVar(&experimentalMode, "experimental", false, "Enables unreleased experimental features")
	rootCmd.PersistentFlags().BoolVar(&silentMode, "silent", false, "Enables log to file intercept.log")
	rootCmd.PersistentFlags().BoolVar(&nologMode, "nolog", false, "Disables all loggging")

	// running id
	intercept_run_id = ksuid.New().String()

	// Setup logging based on verbosity flag
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		setupLogging()
		return setupOutputDir()
	}

}
func setupLogging() {
	switch verbosity {
	case 0:
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case 1:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case 3:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case 4:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	// log.Debug().Msg("This is a debug message")
	// log.Info().Msg("This is an info message")
	// log.Warn().Msg("This is a warning message")
	// log.Error().Msg("This is an error message")
	// log.Fatal().Msg("This is a fatal message and os.Exit(1)")

	// consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	// multi := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)
	// logger := zerolog.New(multi).With().Timestamp().Logger()

	if nologMode {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	// Setup zerolog
	output := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
	log = zerolog.New(output).With().Timestamp().Logger()

	if experimentalMode {
		log = zerolog.New(output).With().Timestamp().Logger().With().Caller().Logger()
		// log = zerolog.New(output).With().Timestamp().Logger().With().Str("id", intercept_run_id).Logger()
	}
	if silentMode {

		logfilepath := fmt.Sprintf("log_intercept_%s.log", intercept_run_id[:6])

		if outputDir != "" {
			logfilepath = filepath.Join(outputDir, logfilepath)
		}

		logFile, _ := lumberjack.NewRoller(
			logfilepath,
			100*1024*1024, // 100 megabytes
			&lumberjack.Options{
				// MaxSize is the maximum size in megabytes of the log file before it gets
				// rotated. It defaults to 100 megabytes.
				MaxBackups: 5,
				// MaxAge is the maximum number of days to retain old log files based on the
				// timestamp encoded in their filename.  Note that a day is defined as 24
				// hours and may not exactly correspond to calendar days due to daylight
				// savings, leap seconds, etc. The default is not to remove old log files
				// based on age.
				MaxAge: 28 * time.Hour * 24, // 28 days
				// Compress determines if the rotated log files should be compressed
				// using gzip. The default is not to perform compression.
				Compress: true,
			})

		if experimentalMode {

			output := zerolog.ConsoleWriter{
				Out:        logFile,
				TimeFormat: time.RFC3339,
				NoColor:    true,
			}
			log = zerolog.New(output).With().Timestamp().Logger().With().Str("intercept_run_id", intercept_run_id).Logger()

		} else {
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
			log = zerolog.New(logFile).With().Timestamp().Logger().With().Str("intercept_run_id", intercept_run_id).Logger()
		}

	}

}

func setupOutputDir() error {
	if outputDir == "" {
		return nil // No output directory specified, nothing to do
	}

	absPath, err := filepath.Abs(outputDir)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get absolute path")
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			log.Fatal().Err(err).Msg("Failed to create output directory")
		}
	} else if err != nil {
		log.Fatal().Err(err).Msg("Failed to stat output directory")
	} else if !info.IsDir() {
		log.Fatal().Msg("Output path is not a directory")
	}

	// Check if the directory is writable
	testFile := filepath.Join(absPath, ".test_write")
	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		log.Fatal().Err(err).Msg("Output directory is not writable")
	}
	os.Remove(testFile) // Clean up the test file

	log.Info().Msgf("Output directory set to: %s", absPath)
	return nil
}

func initConfig() {

	var err error

	rgPath, gossPath, err = prepareEmbeddedExecutables()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to prepare embedded executables")

	} else {

		// Validate the binaries are on the specified paths
		if _, err := os.Stat(rgPath); os.IsNotExist(err) {
			log.Fatal().Msgf("rg executable not found at path: %s", rgPath)
		}
		if _, err := os.Stat(gossPath); os.IsNotExist(err) {
			log.Fatal().Msgf("goss executable not found at path: %s", gossPath)
		}

		// Set the executable permission for the rg binary
		if err := os.Chmod(rgPath, 0755); err != nil {
			log.Debug().Msgf("Failed to set executable permission for rg: %v", err)
			return
		}
		// Set the executable permission for the goss binary
		if err := os.Chmod(gossPath, 0755); err != nil {
			log.Debug().Msgf("Failed to set executable permission for goss: %v", err)
			return
		}
		log.Debug().Msgf("Paths : %s %s", rgPath, gossPath)
	}
}
