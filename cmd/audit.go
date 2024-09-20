package cmd

import (
	"fmt"

	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type Performance struct {
	StartTime time.Time
	EndTime   time.Time
	Delta     time.Duration
}

var (
	targetDir        string
	tagsAny          string
	tagsAll          string
	envDetection     bool
	debugOutput      bool
	rgPath           string
	gossPath         string
	policyFile       string
	policyFileSHA256 string
	outputType       string
	policyData       *PolicyFile
)

var runAuditPerfCmd = &cobra.Command{
	Use:   "audit",
	Short: "Run an optimized audit through all loaded policies",
	Long:  `This command loads all policies from the policy file and runs a parallelized audit.`,
	Run:   runAuditPerf,
}

func init() {
	rootCmd.AddCommand(runAuditPerfCmd)
	runAuditPerfCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Target directory to audit")
	runAuditPerfCmd.Flags().StringVarP(&tagsAny, "tags_any", "f", "", "Filter policies that match any of the provided tags (comma-separated)")
	runAuditPerfCmd.Flags().StringVar(&tagsAll, "tags_all", "", "Filter policies that match all of the provided tags (comma-separated)")
	runAuditPerfCmd.Flags().StringVarP(&environment, "environment", "e", "", "Filter policies that match the specified environment")
	runAuditPerfCmd.Flags().BoolVar(&envDetection, "env-detection", false, "Enable environment detection if no environment is specified")
	runAuditPerfCmd.Flags().BoolVar(&debugOutput, "debug", false, "Enable debug verbose output")
	runAuditPerfCmd.Flags().StringVarP(&policyFile, "policy", "p", "", "policy FILE or URL")
	runAuditPerfCmd.Flags().StringVar(&policyFileSHA256, "checksum", "", "policy file SHA256 checksum")
	runAuditPerfCmd.Flags().StringVar(&outputType, "output", "sarif", "output type")
}

func runAuditPerf(cmd *cobra.Command, args []string) {

	var err error

	perf := Performance{StartTime: time.Now()}

	sourceType, processedInput, err := DeterminePolicySource(policyFile)
	if err != nil {
		log.Fatal().Err(err)
	}

	switch sourceType {
	case LocalFile:
		policyData, err = LoadPolicyFile(processedInput)
	case RemoteURL:
		policyData, err = LoadRemotePolicy(processedInput, policyFileSHA256)
	default:
		log.Fatal().Msg("unknown policy source type")
	}

	//policyData, err = LoadPolicyFile(policyFile)

	if err != nil {
		log.Fatal().Err(err).Str("file", policyFile).Msg("Error loading policy file")
	}

	// Clean up output directories
	if err := cleanupOutputDirectories(); err != nil {
		log.Fatal().Err(err).Msg("Failed to clean up output directories")

	}

	if err := initSARIFProcessing(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize SARIF processing")

	}
	defer cleanupSARIFProcessing()

	if err := createOutputDirectories(false); err != nil {
		log.Fatal().Err(err).Msg("Failed to create output directories")

	}

	config := GetConfig()

	policies_provided := GetPolicies()
	policies_filtered := filterPolicies(policies_provided, config.Flags.Tags)

	// Override output type if provided in command line
	if outputType != "" {
		policyData.Config.Flags.OutputType = outputType
	}

	// printConfigInfo(config, policies_filtered, environment, rgPath, gossPath)

	log.Info().Msgf("Total Policies: %d", len(policies_provided))
	log.Info().Msgf("Total Policies (after filtering): %d", len(policies_filtered))

	// Check if all policies are API type
	allAPIorRuntimePolicies := true
	for _, policy := range policies_filtered {
		if policy.Type != "api" && policy.Type != "runtime" {
			allAPIorRuntimePolicies = false
			break
		}
	}

	if targetDir == "" {
		targetDir = config.Flags.Target
	}

	if !allAPIorRuntimePolicies {
		if targetDir == "" {
			log.Fatal().Msg("Target directory is required for non-API/Runtime policies")

		}
		log.Debug().Msgf("Processing target directory: %s", targetDir)

		allFileInfos, err := CalculateFileHashes(targetDir)

		if err != nil {
			log.Err(err).Msg("Error verifying target")
		} else {
			processPoliciesInParallel(policies_filtered, allFileInfos, rgPath)
		}
	} else {
		log.Info().Msg("All policies are API or Runtime type, skipping target directory processing")
		processPoliciesInParallel(policies_filtered, nil, rgPath)
	}

	perf.EndTime = time.Now()
	perf.Delta = perf.EndTime.Sub(perf.StartTime)

	commandLine := strings.Join(os.Args, " ")

	_, err = MergeSARIFReports(commandLine, perf, false)

	if err != nil {
		log.Debug().Err(err).Msg("Failed to merge SARIF reports")
	}

	log.Info().Msgf("INTERCEPT Run ID: %s", intercept_run_id)
	log.Info().Msg("Performance Metrics:")
	log.Info().Msgf("  Start Time: %s", perf.StartTime.Format(time.RFC3339))
	log.Info().Msgf("  End Time: %s", perf.EndTime.Format(time.RFC3339))
	log.Info().Msgf("  Execution Time: %d milliseconds", perf.Delta.Milliseconds())

}

func filterPolicies(policies []Policy, config_tags []string) []Policy {
	if environment == "" && envDetection {
		environment = DetectEnvironment()
	}

	filtered := policies

	if len(config_tags) > 0 {
		filtered = FilterPoliciesByAnyTags(filtered, config_tags)
	}

	if tagsAny != "" {
		filtered = FilterPoliciesByAnyTags(filtered, strings.Split(tagsAny, ","))
	}

	if tagsAll != "" {
		filtered = FilterPoliciesByAllTags(filtered, strings.Split(tagsAll, ","))
	}

	if environment != "" {
		filtered = FilterPoliciesByEnvironment(filtered, environment)
	}

	return filtered
}

func processPoliciesInParallel(policies []Policy, allFileInfos []FileInfo, rgPath string) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 100) // Limit concurrent goroutines

	for _, policy := range policies {
		wg.Add(1)
		go func(p Policy) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			processPolicy(p, allFileInfos, rgPath)
		}(policy)
	}

	wg.Wait()
}

func processPolicy(policy Policy, allFileInfos []FileInfo, rgPath string) {
	filesToProcess, filePaths := filterFiles(policy, allFileInfos)

	if policy.Type == "json" || policy.Type == "yaml" || policy.Type == "ini" || policy.Type == "scan" || policy.Type == "assure" {

		log.Debug().Str("policy", policy.ID).Msgf(" Processing files for policy %s ", policy.ID)
		if len(filesToProcess) < 15 {
			for _, file := range filesToProcess {
				log.Debug().Str("policy", policy.ID).Msgf("  %s: %s ", file.Path, file.Hash)
			}
		}

		normalizedID := NormalizeFilename(policy.ID)
		outputPath := fmt.Sprintf("scanned_files_%s.json", normalizedID)

		err := WriteHashesToJSON(filesToProcess, outputPath)
		if err != nil {
			log.Debug().Msgf("Error writing hashes to JSON for policy %s: %v ", policy.ID, err)
		} else {
			log.Debug().Msgf("File hashes for policy %s written to: %s ", policy.ID, outputPath)
		}
	}
	processPolicyByType(policy, rgPath, gossPath, targetDir, filePaths)
}

func filterFiles(policy Policy, allFileInfos []FileInfo) ([]FileInfo, []string) {
	var filesToProcess []FileInfo
	var filePaths []string

	if policy.FilePattern != "" {
		filteredFiles, err := FilterFilesByPattern(allFileInfos, policy.FilePattern)
		if err != nil {
			log.Debug().Msgf("Error filtering files for policy %s: %v ", policy.ID, err)
			return filesToProcess, filePaths
		}
		filesToProcess = filteredFiles
	} else {
		filesToProcess = allFileInfos
	}

	for _, file := range filesToProcess {
		filePaths = append(filePaths, file.Path)
	}

	return filesToProcess, filePaths
}

func processPolicyByType(policy Policy, rgPath, gossPath, targetDir string, filePaths []string) {
	var err error
	switch policy.Type {
	case "scan":
		err = ProcessScanType(policy, rgPath, targetDir, filePaths)
	case "assure":
		err = ProcessAssureType(policy, rgPath, targetDir, filePaths)
	case "runtime":
		err = ProcessRuntimeType(policy, gossPath, targetDir, filePaths, false)
	case "api":
		err = ProcessAPIType(policy, rgPath)
	case "yml":
		if policy.Schema.Patch {
			err = processGenericType(policy, filePaths, "yaml")
		} else {
			err = ProcessYAMLType(policy, targetDir, filePaths)
		}
	case "toml":
		if policy.Schema.Patch {
			err = processGenericType(policy, filePaths, "toml")
		} else {
			err = ProcessTOMLType(policy, targetDir, filePaths)
		}
	case "json":
		if policy.Schema.Patch {
			err = ProcessJSONTypeWithPatch(policy, targetDir, filePaths)
		} else {
			err = ProcessJSONType(policy, targetDir, filePaths)
		}
	case "ini":
		if policy.Schema.Patch {
			err = processGenericType(policy, filePaths, "ini")
		} else {
			err = ProcessINIType(policy, targetDir, filePaths)
		}
	case "rego":
		err = ProcessRegoType(policy, targetDir, filePaths)
	default:
		log.Debug().Msgf("Unsupported policy type %s for policy %s ", policy.Type, policy.ID)
		return
	}

	if err != nil {
		log.Debug().Msgf("Error processing %s-type policy %s: %v ", policy.Type, policy.ID, err)
	}
}

func GetConfig() Config {
	return policyData.Config
}

func GetPolicies() []Policy {
	return policyData.Policies
}
