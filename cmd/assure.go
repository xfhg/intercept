package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// ProcessAssureType handles the assurance process for policies of type "assure"
func ProcessAssureType(policy Policy, rgPath string, targetDir string, filePaths []string) error {
	if policy.Type == "assure" {
		err := executeAssure(policy, rgPath, targetDir, filePaths)
		if err != nil {
			log.Error().Err(err).Msgf("error assuring for policy %s", policy.ID)
			return fmt.Errorf("error assuring for policy %s: %w", policy.ID, err)
		}
	}

	return nil
}

func executeAssure(policy Policy, rgPath string, targetDir string, filesToAssure []string) error {
	// Create a temporary file to store the search patterns
	searchPatternFile, err := createSearchPatternFile(policy.Regex, NormalizeFilename(policy.ID))
	if err != nil {
		log.Error().Err(err).Msg("error creating search pattern file")
		return fmt.Errorf("error creating search pattern file: %w", err)
	}
	defer os.Remove(searchPatternFile) // Clean up the temporary file

	// Prepare the ripgrep command for JSON output
	jsonOutputFile := fmt.Sprintf("assure_results_%s.json", NormalizeFilename(policy.ID))

	if outputDir != "" {
		jsonOutputFile = filepath.Join(outputDir, "_debug", jsonOutputFile)
	} else {
		jsonOutputFile = filepath.Join("_debug", jsonOutputFile)
	}

	jsonoutfile, err := os.Create(jsonOutputFile)
	if err != nil {
		log.Error().Err(err).Msg("error creating JSON output file")
		return fmt.Errorf("error creating JSON output file: %w", err)
	}
	defer jsonoutfile.Close()

	writer := bufio.NewWriter(jsonoutfile)
	defer writer.Flush()

	processedIgnorePatterns := processIgnorePatterns(policyData.Config.Flags.Ignore)

	codePatternAssureJSON := []string{
		"--pcre2",
		"--no-heading",
		"-o",
		"-p",
		"-i",
		"-U",
		"--json",
		"-f", searchPatternFile,
	}

	if targetDir == "" {
		return fmt.Errorf("no target directory defined")
	}

	codePatternAssureJSON = append(codePatternAssureJSON, processedIgnorePatterns...)

	matchesFound := true

	// Parallel execution for large file sets
	if policy.FilePattern == "" {
		log.Warn().Str("policy", policy.ID).Msg("ASSURE Policy without a filepattern is suboptimal")
	}
	if len(filesToAssure) > 25 {
		matchesFound, err = executeParallelAssure(rgPath, codePatternAssureJSON, filesToAssure, writer)
	} else {
		matchesFound, err = executeSingleAssure(rgPath, codePatternAssureJSON, filesToAssure, targetDir, policy, writer)
	}

	if err != nil {

		log.Error().Err(err).Msg("error executing ripgrep batch")

	}

	// Patch the JSON output file
	err = patchJSONOutputFile(jsonOutputFile)
	if err != nil {
		log.Error().Err(err).Msg("error patching JSON output file")
		return fmt.Errorf("error patching JSON output file: %w", err)
	}

	log.Debug().Msgf("JSON output for assure policy %s written to: %s ", policy.ID, jsonOutputFile)
	log.Debug().Msgf("Scanned ~%d files for policy %s", len(filesToAssure), policy.ID)

	// Determine the status based on whether matches were found
	status := "NOT FOUND"
	if matchesFound {
		status = "FOUND"
	}

	// Generate SARIF report
	sarifReport, err := GenerateAssureSARIFReport(jsonOutputFile, policy, status)
	if err != nil {
		log.Error().Err(err).Msg("error generating SARIF report")
		return fmt.Errorf("error generating SARIF report: %w", err)
	}

	// Write SARIF report to file
	var sarifOutputFile string

	if policy.RunID != "" {
		if err := writeSARIFReport(policy.RunID, sarifReport); err != nil {
			log.Error().Err(err).Msg("error writing SARIF report")
			return fmt.Errorf("error writing SARIF report: %w", err)
		}
		sarifOutputFile = fmt.Sprintf("%s.sarif", policy.RunID)
	} else {
		if err := writeSARIFReport(policy.ID, sarifReport); err != nil {
			log.Error().Err(err).Msg("error writing SARIF report")
			return fmt.Errorf("error writing SARIF report: %w", err)
		}
		sarifOutputFile = fmt.Sprintf("%s.sarif", NormalizeFilename(policy.ID))

	}

	log.Debug().Msgf("Policy %s processed. SARIF report written to: %s ", policy.ID, sarifOutputFile)

	return nil
}

func executeParallelAssure(rgPath string, baseArgs []string, filesToScan []string, writer *bufio.Writer) (bool, error) {

	const batchSize = 25
	matched := true
	var wg sync.WaitGroup
	errChan := make(chan error, len(filesToScan)/batchSize+1)
	var mu sync.Mutex

	for i := 0; i < len(filesToScan); i += batchSize {
		end := i + batchSize
		if end > len(filesToScan) {
			end = len(filesToScan)
		}
		batch := filesToScan[i:end]

		// log.Debug().Msgf("RGM: %v", batch)

		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()
			args := append(baseArgs, batch...)
			cmd := exec.Command(rgPath, args...)
			output, err := cmd.Output()

			if err != nil {

				if exitError, ok := err.(*exec.ExitError); ok {
					// Exit code 1 in ripgrep means "no matches found"
					if exitError.ExitCode() == 1 {
						matched = false
						err = nil // Reset error as this is the expected outcome for assure
					}
				}

				if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() != 1 {
					errChan <- fmt.Errorf("error executing ripgrep: %w", err)
					return
				}
			}

			mu.Lock()
			_, writeErr := writer.Write(output)
			if writeErr == nil {
				writeErr = writer.Flush()
			}
			mu.Unlock()

			if writeErr != nil {
				errChan <- fmt.Errorf("error writing output: %w", writeErr)
			}
		}(batch)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return matched, err
		}
	}

	return matched, nil
}

func executeSingleAssure(rgPath string, baseArgs []string, filesToScan []string, targetDir string, policy Policy, writer *bufio.Writer) (bool, error) {

	if len(filesToScan) > 0 {
		baseArgs = append(baseArgs, filesToScan...)
	} else {
		log.Error().Str("policy", policy.ID).Msgf("no files matched policy pattern on target : %s", targetDir)
	}

	matched := true

	// log.Debug().Msgf("RGS: %v", baseArgs)

	cmdJSON := exec.Command(rgPath, baseArgs...)
	cmdJSON.Stdout = writer
	cmdJSON.Stderr = os.Stderr

	log.Debug().Msgf("Creating JSON output for policy %s... ", policy.ID)
	err := cmdJSON.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {

			if exitError.ExitCode() == 1 {
				matched = false
				err = nil // Reset error as this is the expected outcome for assure
			}

			if exitError.ExitCode() == 2 {
				log.Warn().Msgf("RG exited with code 2")
				log.Debug().Msgf("RG Error Args: %v", baseArgs)
				if len(exitError.Stderr) > 0 {
					log.Debug().Msgf("RG exited with code 2 stderr: %s", string(exitError.Stderr))
				}
			}
			if exitError.ExitCode() != 1 {
				log.Error().Err(err).Msg("error executing ripgrep for JSON output")
				return matched, fmt.Errorf("error executing ripgrep for JSON output: %w", err)
			}

		} else {
			log.Error().Err(err).Msg("error executing ripgrep for JSON output")
			return matched, fmt.Errorf("error executing ripgrep for JSON output: %w", err)
		}
	}

	return matched, nil
}
