//go:build !windows
// +build !windows

package cmd

import (
	"context"
	"fmt"

	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/adhocore/gronx"
	"github.com/adhocore/gronx/pkg/tasker"
	"github.com/segmentio/ksuid"
	"github.com/spf13/cobra"
)

var (
	observeTagsAny      string
	observeTagsAll      string
	observeEnvironment  string
	observeEnvDetection bool
	observeDebugOutput  bool
	observePolicyFile   string
	observeSchedule     string
	observeReport       string
	reportMutex         sync.Mutex
	reportDir           string = "_status"
	allFileInfos        []FileInfo
)

var observeCmd = &cobra.Command{
	Use:   "observe",
	Short: "Observe and trigger realtime policies based on schedules or active path monitoring",
	Long:  `This command loads your intercept policies and runs them based on their schedules or triggers them from path/file monitoring`,
	Run:   runObserve,
}

func init() {
	rootCmd.AddCommand(observeCmd)

	observeCmd.Flags().StringVar(&observeTagsAny, "tags_any", "", "Filter policies that match any of the provided tags (comma-separated)")
	observeCmd.Flags().StringVar(&observeTagsAll, "tags_all", "", "Filter policies that match all of the provided tags (comma-separated)")
	observeCmd.Flags().StringVar(&observeEnvironment, "environment", "", "Filter policies that match the specified environment")
	observeCmd.Flags().BoolVar(&observeEnvDetection, "env-detection", false, "Enable environment detection if no environment is specified")
	observeCmd.Flags().BoolVar(&observeDebugOutput, "debug", false, "Enable debug verbose output")
	observeCmd.Flags().StringVar(&observePolicyFile, "policy", "", "Policy file")
	observeCmd.Flags().StringVar(&observeSchedule, "schedule", "", "Global Cron Schedule")
	observeCmd.Flags().StringVar(&observeReport, "report", "", "Report Cron Schedule")

}

func runObserve(cmd *cobra.Command, args []string) {

	perf := Performance{StartTime: time.Now()}

	if observePolicyFile != "" {
		var err error
		policyData, err = LoadPolicyFile(observePolicyFile)
		if err != nil {
			log.Fatal().Err(err).Str("file", observePolicyFile).Msg("Error loading policy file")
		}
	}
	// Clean up output directories
	if err := cleanupOutputDirectories(); err != nil {
		log.Fatal().Err(err).Msg("Failed to clean up output directories")

	}

	if err := initSARIFProcessing(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize SARIF processing")

	}
	defer cleanupSARIFProcessing()

	if err := createOutputDirectories(true); err != nil {
		log.Fatal().Err(err).Msg("Failed to create output directories")

	}
	if outputDir != "" {
		reportDir = filepath.Join(outputDir, reportDir)
		log.Debug().Msgf("Setting up report directory: %s", reportDir)
	}
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		log.Fatal().Err(err).Msgf("failed to create directory %s", reportDir)
	}

	config := GetConfig()

	// Needed for scan/assure/schema policies
	if config.Flags.Target != "" {
		targetDir = config.Flags.Target
		allFileInfos, _ = CalculateFileHashes(targetDir)
		log.Debug().Msgf("Setting up policy target directory: %s", targetDir)
	}

	// check embed
	log.Debug().Msgf("Embedded X: %s", rgPath)
	log.Debug().Msgf("Embedded Y: %s", gossPath)

	// check if global schedule
	if observeSchedule != "" {
		config.Flags.PolicySchedule = observeSchedule
	}

	// check if schedule Report
	if observeReport != "" {
		config.Flags.ReportSchedule = observeReport
	}

	policies := loadFilteredPolicies()
	if len(policies) == 0 {
		log.Fatal().Msg("No active policies found")
		return
	}

	dispatcher := GetDispatcher()

	taskr := tasker.New(tasker.Option{
		Verbose: observeDebugOutput,
		Tz:      "UTC", // You can change this to your preferred timezone
	})

	for _, policy := range policies {

		// SCHEDULERS
		schedule := getScheduleForPolicy(policy, config.Flags.PolicySchedule)
		if schedule == "" {
			log.Warn().Str("policy", policy.ID).Msg("No schedule found for policy, skipping")
			continue
		}

		if !validateCronExpression(schedule) {
			log.Error().Str("policy", policy.ID).Str("schedule", schedule).Msg("Invalid cron expression, skipping")
			continue
		}
		if policy.Type != "api" && policy.Type != "runtime" && policy.Type != "rego" {
			policy.Metadata.TargetInfo = preparePolicyPaths(policy, allFileInfos)
		}

		policyTask := createPolicyTask(policy, dispatcher)
		taskr.Task(schedule, policyTask)
		log.Info().Str("policy", policy.ID).Str("schedule", schedule).Msg("Added policy to Scheduler")

		//PATH WATCHERS
		if policy.Type == "runtime" && policy.Runtime.Observe != "" {

			exists, isDirectory, _ := PathInfo(policy.Runtime.Observe)

			if exists {

				log.Debug().Str("policy", policy.ID).Bool("exists", exists).Bool("isDirectory", isDirectory).Msgf("Setting up watch : %s", policy.Runtime.Observe)

				StorePolicyInCache(policy.Runtime.Observe, policy)

				log.Debug().Int("Cache count", GetPolicyCacheCount()).Msg("Cache Status")

				if PolicyExistsInCache(policy.Runtime.Observe) {
					log.Info().Str("policy", policy.ID).Str("Observe", policy.Runtime.Observe).Msg("Added policy to Path Watcher")

					//path watcher
					go func() {
						defer func() {
							if r := recover(); r != nil {
								log.Error().Interface("recover", r).Msg("Panic in path watcher goroutine")
							}
						}()
						watchPaths(policy.Runtime.Observe)
					}()

				} else {
					log.Warn().Str("policy", policy.ID).Msg("Failed Caching the policy - investigate")
				}

			} else {
				log.Warn().Str("policy", policy.ID).Msg("Runtime observe has invalid path, skipping")
			}

		}

	}

	if config.Flags.ReportSchedule != "" {
		if validateCronExpression(config.Flags.ReportSchedule) {
			reportTask := createReportTask()
			taskr.Task(config.Flags.ReportSchedule, reportTask)
			log.Info().Str("schedule", config.Flags.ReportSchedule).Msg("Added Report Task to Scheduler")
		} else {
			log.Fatal().Str("schedule", config.Flags.ReportSchedule).Msg("Invalid cron expression for Report, quitting")
		}
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Info().Msg("Received termination signal. Initiating graceful shutdown...")
		cancel()
	}()

	log.Info().Msg("Observing...")
	go taskr.Run()

	<-ctx.Done()
	observeCleanup(perf)

}

func loadFilteredPolicies() []Policy {
	policies := GetPolicies()
	return ofilterPolicies(policies)
}

func ofilterPolicies(policies []Policy) []Policy {
	if observeEnvironment == "" && observeEnvDetection {
		observeEnvironment = DetectEnvironment()
	}

	filtered := policies

	if observeTagsAny != "" {
		filtered = FilterPoliciesByAnyTags(filtered, strings.Split(observeTagsAny, ","))
	}

	if observeTagsAll != "" {
		filtered = FilterPoliciesByAllTags(filtered, strings.Split(observeTagsAll, ","))
	}

	if observeEnvironment != "" {
		filtered = FilterPoliciesByEnvironment(filtered, observeEnvironment)
	}

	return filtered
}

func getScheduleForPolicy(policy Policy, globalSchedule string) string {
	if policy.Schedule != "" {
		return policy.Schedule
	}
	return globalSchedule
}

func createPolicyTask(policy Policy, dispatcher *Dispatcher) func(context.Context) (int, error) {
	return func(ctx context.Context) (int, error) {
		runID := fmt.Sprintf("%s-%s", ksuid.New().String(), NormalizeFilename(policy.ID))
		log.Info().Str("policy", policy.ID).Str("runID", runID).Msg("Executing policy")

		// Set the RunID for the policy
		policy.RunID = runID

		err := dispatcher.DispatchPolicyEvent(policy, targetDir, policy.Metadata.TargetInfo)
		if err != nil {
			log.Error().Err(err).Str("policy", policy.ID).Str("runID", runID).Msg("Failed to execute policy")
			return 1, err
		}

		log.Info().Str("policy", policy.ID).Str("runID", runID).Msg("Successfully executed policy")
		return 0, nil
	}
}

func createReportTask() func(context.Context) (int, error) {
	return func(ctx context.Context) (int, error) {
		scheduledReport()
		log.Info().Msg("Executed Report Task")
		return 0, nil
	}
}

func validateCronExpression(expr string) bool {
	gron := gronx.New()
	return gron.IsValid(expr)
}

func observeCleanup(perf Performance) {
	log.Warn().Msg("Performing cleanup tasks... (5 seconds)")

	time.Sleep(3 * time.Second) // Give some time for tasks to complete

	reportMutex.Lock()
	defer reportMutex.Unlock()

	commandLine := strings.Join(os.Args, " ")

	mergedReport, err := MergeSARIFReports(commandLine, perf, true)

	if err != nil {
		log.Debug().Err(err).Msg("Failed to merge SARIF reports")
	}

	if len(mergedReport.Runs) == 0 {
		log.Warn().Msg("Merged SARIF report contains no runs")
	} else {
		// Post merged SARIF report to webhooks
		if err := PostReportToWebhooks(mergedReport); err != nil {
			log.Error().Err(err).Msg("Failed to post merged SARIF report to webhooks")
		}
	}

	if err := manageStatusReports(); err != nil {
		log.Error().Err(err).Msg("Failed to manage status reports")
	}

	perf.EndTime = time.Now()
	perf.Delta = perf.EndTime.Sub(perf.StartTime)

	// Clean up the _sarif folder
	if err := cleanupSARIFFolder(); err != nil {
		log.Error().Err(err).Msg("Failed to clean up SARIF folder")
	}

	log.Info().Msg("Performance Metrics:")
	log.Info().Msgf("  Start Time: %s", perf.StartTime.Format(time.RFC3339))
	log.Info().Msgf("  End Time: %s", perf.EndTime.Format(time.RFC3339))
	log.Info().Msgf("  Execution Time: %d milliseconds", perf.Delta.Milliseconds())

	time.Sleep(2 * time.Second)

	log.Info().Msg("Cleanup completed. Exiting...")
}

func scheduledReport() {
	log.Info().Msg("Scheduled Report Generation...")

	reportMutex.Lock()
	defer reportMutex.Unlock()

	commandLine := strings.Join(os.Args, " ")
	perf := Performance{StartTime: time.Now()}
	perf.EndTime = time.Now().Add(time.Second)
	perf.Delta = perf.EndTime.Sub(perf.StartTime)

	mergedReport, err := MergeSARIFReports(commandLine, perf, true)
	if err != nil {
		log.Error().Err(err).Msg("Failed to merge SARIF reports")
		return
	}

	if len(mergedReport.Runs) == 0 {
		log.Warn().Msg("Merged SARIF report contains no runs")
	} else {
		// Post merged SARIF report to webhooks
		if err := PostReportToWebhooks(mergedReport); err != nil {
			log.Error().Err(err).Msg("Failed to post merged SARIF report to webhooks")
		}
	}

	if err := manageStatusReports(); err != nil {
		log.Error().Err(err).Msg("Failed to manage status reports")
	}

	// Clean up the _sarif folder
	if err := cleanupSARIFFolder(); err != nil {
		log.Error().Err(err).Msg("Failed to clean up SARIF folder")
	}

	log.Info().Msg("Scheduled report generation completed.")
}
