package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	searchPatternFile, err := createSearchPatternFile(policy.Regex)
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

	// Append the file targets
	if len(filesToAssure) > 0 {
		codePatternAssureJSON = append(codePatternAssureJSON, filesToAssure...)
	} else if policy.FilePattern == "" {
		codePatternAssureJSON = append(codePatternAssureJSON, targetDir)
	} else {
		return fmt.Errorf("no files matched policy pattern")
	}

	// Execute the ripgrep command for JSON output
	cmdJSON := exec.Command(rgPath, codePatternAssureJSON...)
	cmdJSON.Stdout = writer
	cmdJSON.Stderr = os.Stderr

	log.Debug().Msgf("Creating JSON output for assure policy %s... ", policy.ID)
	err = cmdJSON.Run()

	// Check if ripgrep found any matches
	matchesFound := true
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 1 in ripgrep means "no matches found"
			if exitError.ExitCode() == 1 {
				matchesFound = false
				err = nil // Reset error as this is the expected outcome for assure
			} else {
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

	log.Debug().Msgf("JSON output for assure policy %s written to: %s ", policy.ID, jsonOutputFile)

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
