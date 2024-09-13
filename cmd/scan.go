package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProcessScanType handles the scanning process for policies of type "scan"
func ProcessScanType(policy Policy, rgPath string, targetDir string, filePaths []string) error {
	if policy.Type == "scan" {
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
	searchPatternFile, err := createSearchPatternFile(policy.Regex)
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

	if targetDir == "" {
		return fmt.Errorf("no target directory defined")
	}

	// Append the same file targets as the previous command
	if len(filesToScan) > 0 {
		codePatternScanJSON = append(codePatternScanJSON, filesToScan...)
	} else if policy.FilePattern == "" {
		codePatternScanJSON = append(codePatternScanJSON, targetDir)
	} else {
		return fmt.Errorf("no files matched policy pattern")
	}

	// Execute the ripgrep command for JSON output
	cmdJSON := exec.Command(rgPath, codePatternScanJSON...)
	cmdJSON.Stdout = writer
	cmdJSON.Stderr = os.Stderr

	log.Debug().Msgf("Creating JSON output for policy %s... ", policy.ID)
	err = cmdJSON.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 1 in ripgrep means "no matches found", which isn't an error for us
			if exitError.ExitCode() != 1 {
				log.Error().Err(err).Msg("error executing ripgrep for JSON output")
				return fmt.Errorf("error executing ripgrep for JSON output: %w", err)
			}
		} else {
			log.Error().Err(err).Msg("error executing ripgrep for JSON output")
			return fmt.Errorf("error executing ripgrep for JSON output: %w", err)
		}
	}

	// Patch the JSON output file
	err = patchJSONOutputFile(jsonOutputFile)
	if err != nil {
		log.Error().Err(err).Msg("error patching JSON output file")
		return fmt.Errorf("error patching JSON output file: %w", err)
	}

	log.Debug().Msgf("JSON output for policy %s written to: %s ", policy.ID, jsonOutputFile)

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

func createSearchPatternFile(patterns []string) (string, error) {
	searchPattern := []byte(strings.Join(patterns, "\n") + "\n")

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "policypatterns")
	if err != nil {
		return "", fmt.Errorf("error creating temporary file: %w", err)
	}

	// Write the search patterns to the file
	if _, err := tmpfile.Write(searchPattern); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		return "", fmt.Errorf("error writing to temporary file: %w", err)
	}

	if err := tmpfile.Close(); err != nil {
		os.Remove(tmpfile.Name())
		return "", fmt.Errorf("error closing temporary file: %w", err)
	}

	return tmpfile.Name(), nil
}
