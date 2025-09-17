package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ProcessScanType handles the scanning process for policies of type "scan"
func ProcessScanType(policy Policy, rgPath string, targetDir string, filePaths []string) error {
	if policy.Type == "scan" {
		if !featureRgReady || rgPath == "" {
			return fmt.Errorf("rg executable unavailable: scan policies cannot be processed")
		}
		err := executeScan(policy, rgPath, targetDir, filePaths)
		if err != nil {
			log.Error().Str("Policy: ", policy.ID).Msg("Error Scanning Policy")
			log.Error().Err(err).Msg("error scanning for policy %s")
			return fmt.Errorf("error scanning for policy %s: %w", policy.ID, err)
		}
	}

	return nil
}

func executeScan(policy Policy, rgPath string, targetDir string, filesToScan []string) error {
	// Create a temporary file to store the search patterns
	searchPatternFile, err := createSearchPatternFile(policy.Regex, NormalizeFilename(policy.ID))
	if err != nil {
		log.Error().Err(err).Msg("error creating search pattern file")
		return fmt.Errorf("error creating search pattern file: %w", err)
	}
	defer os.Remove(searchPatternFile) // Clean up the temporary file

	// Prepare the ripgrep command for JSON output
	jsonOutputFile := fmt.Sprintf("scan_results_%s.json", NormalizeFilename(policy.ID))

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

	codePatternScanJSON := []string{
		"--pcre2",
		"--no-heading",
		"-o",
		"-p",
		"-i",
		"-U",
		"--json",
		"-f", searchPatternFile,
	}
	codePatternScanJSON = append(codePatternScanJSON, processedIgnorePatterns...)

	if targetDir == "" {
		return fmt.Errorf("no target directory defined")
	}

	// Parallel execution for large file sets
	if policy.FilePattern == "" {
		err = executeSingleScan(rgPath, codePatternScanJSON, nil, targetDir, policy, writer)
	} else if len(filesToScan) > 25 {
		err = executeParallelScans(rgPath, codePatternScanJSON, filesToScan, writer)
	} else {
		err = executeSingleScan(rgPath, codePatternScanJSON, filesToScan, targetDir, policy, writer)
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

	log.Debug().Msgf("JSON output for policy %s written to: %s ", policy.ID, jsonOutputFile)
	log.Debug().Msgf("Scanned ~%d files for policy %s", len(filesToScan), policy.ID)

	// Generate SARIF report
	sarifReport, err := GenerateSARIFReport(jsonOutputFile, policy)
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

func createSearchPatternFile(patterns []string, policyId string) (string, error) {
	searchPattern := []byte(strings.Join(patterns, "\n") + "\n")

	filename := filepath.Join("_debug", policyId)
	if outputDir != "" {
		filename = filepath.Join(outputDir, filename)
	}

	tmpfile, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("error creating file in _debug directory: %w", err)
	}

	// Write the search patterns to the file
	if _, err := tmpfile.Write(searchPattern); err != nil {
		tmpfile.Close()
		os.Remove(filename)
		return "", fmt.Errorf("error writing to file: %w", err)
	}

	if err := tmpfile.Close(); err != nil {
		os.Remove(filename)
		return "", fmt.Errorf("error closing file: %w", err)
	}

	return tmpfile.Name(), nil
}

func executeParallelScans(rgPath string, baseArgs []string, filesToScan []string, writer *bufio.Writer) error {
	const batchSize = 25
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
			return err
		}
	}

	return nil
}

func executeSingleScan(rgPath string, baseArgs []string, filesToScan []string, targetDir string, policy Policy, writer *bufio.Writer) error {
	if len(filesToScan) > 0 {
		baseArgs = append(baseArgs, filesToScan...)
	} else if policy.FilePattern == "" {
		baseArgs = append(baseArgs, targetDir)
	} else {
		log.Error().Str("policy", policy.ID).Msgf("no files matched policy pattern on target : %s", targetDir)
	}

	// log.Debug().Msgf("RGS: %v", baseArgs)

	cmdJSON := exec.Command(rgPath, baseArgs...)
	cmdJSON.Stdout = writer
	cmdJSON.Stderr = os.Stderr

	log.Debug().Msgf("Creating JSON output for policy %s... ", policy.ID)
	err := cmdJSON.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 2 {
				log.Warn().Msgf("RG exited with code 2")
				log.Debug().Msgf("RG Error Args: %v", baseArgs)
				if len(exitError.Stderr) > 0 {
					log.Debug().Msgf("RG exited with code 2 stderr: %s", string(exitError.Stderr))
				}
			}
			if exitError.ExitCode() != 1 {
				log.Error().Err(err).Msg("error executing ripgrep for JSON output")
				return fmt.Errorf("error executing ripgrep for JSON output: %w", err)
			}

		} else {
			log.Error().Err(err).Msg("error executing ripgrep for JSON output")
			return fmt.Errorf("error executing ripgrep for JSON output: %w", err)
		}
	}

	return nil
}
