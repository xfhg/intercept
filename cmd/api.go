package cmd

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"

	"github.com/go-resty/resty/v2"
)

func ProcessAPIType(policy Policy, rgPath string) error {
	client := resty.New()
	client.SetDebug(debugOutput)

	if policy.API.Insecure {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	req := client.R()

	req.SetHeader("Content-Type", policy.API.ResponseType)
	req.SetHeader("User-Agent", "INTERCEPT/v1.0.X")

	// Apply authentication
	if err := applyAuth(req, policy.API.Auth); err != nil {
		log.Error().Err(err).Msg("error applying authentication")
		return fmt.Errorf("error applying authentication: %w", err)
	}

	// Set request body if method is POST
	if policy.API.Method == "POST" && policy.API.Body != "" {
		req.SetBody(policy.API.Body)
	}

	// Make the API request
	resp, err := req.Execute(policy.API.Method, policy.API.Endpoint)
	if err != nil {
		log.Error().Err(err).Msg("error making API request")
		return fmt.Errorf("error making API request: %w", err)
	}

	// check for accepted policy.API.ResponseType and map to schema type is defined , or use regex

	// Process the response based on policy type
	if policy.Schema.Structure != "" {
		return processWithCUE(policy, resp.Body())
	} else if len(policy.Regex) > 0 {
		return processWithRegex(policy, resp.Body(), rgPath)
	}

	return fmt.Errorf("no processing method specified for policy %s", policy.ID)
}

func applyAuth(req *resty.Request, auth map[string]string) error {
	authType, ok := auth["type"]
	if !ok {
		return nil // No authentication specified
	}

	switch authType {
	case "basic":
		username := os.Getenv(auth["username_env"])
		password := os.Getenv(auth["password_env"])
		req.SetHeader("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
	case "bearer":
		token := os.Getenv(auth["token_env"])
		req.SetHeader("Authorization", "Bearer "+token)
	case "api_key":
		key := os.Getenv(auth["key_env"])
		req.SetHeader(auth["header"], key)
	default:
		return fmt.Errorf("unsupported authentication type: %s", authType)
	}

	return nil
}

func processWithCUE(policy Policy, data []byte) error {
	valid, issues := validateContentAndCUE(data, policy.Schema.Structure, "json", policy.Schema.Strict, policy.ID)

	// Generate SARIF report
	sarifReport, err := GenerateAPISARIFReport(policy, policy.API.Endpoint, valid, issues)
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

	if !valid {
		log.Debug().Msgf("Policy %s validation failed for API response: ", policy.ID)
		for _, issue := range issues {
			log.Debug().Msgf("- %s ", issue)
		}
		return fmt.Errorf("API response failed validation")
	}
	log.Debug().Msgf("Policy %s validation passed for API response ", policy.ID)
	return nil
}
func processWithRegex(policy Policy, data []byte, rgPath string) error {
	// Create a temporary file with the API response
	tempFile, err := os.CreateTemp("", "api_response_*.json")
	if err != nil {
		log.Error().Err(err).Msg("error creating temporary file")
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write(data); err != nil {
		log.Error().Err(err).Msg("error writing API response to temporary file")
		return fmt.Errorf("error writing API response to temporary file: %w", err)
	}
	tempFile.Close()

	// Execute assure and get the results
	matchesFound, err := executeAssureForAPI(policy, rgPath, tempFile.Name())
	if err != nil {
		log.Error().Err(err).Msg("error executing assure for API")
		return fmt.Errorf("error executing assure for API: %w", err)
	}

	var issues []string
	if !matchesFound {
		issues = append(issues, "Required pattern not found in API response")
	}

	// Generate SARIF report
	sarifReport, err := GenerateAPISARIFReport(policy, policy.API.Endpoint, matchesFound, issues)
	if err != nil {
		log.Error().Err(err).Msg("error generating SARIF report")
		return fmt.Errorf("error generating SARIF report: %w", err)
	}

	// Write SARIF report
	err = writeSARIFReport(policy.ID, sarifReport)
	if err != nil {
		log.Error().Err(err).Msg("error writing SARIF report for policy %s")
		return fmt.Errorf("error writing SARIF report for policy %s: %w", policy.ID, err)
	}

	if matchesFound {
		log.Debug().Msgf("Policy %s assurance passed for API response (pattern found) ", policy.ID)
	} else {
		log.Debug().Msgf("Policy %s assurance failed for API response (pattern not found) ", policy.ID)
		return fmt.Errorf("API response failed assurance check")
	}
	return nil
}
func executeAssureForAPI(policy Policy, rgPath, filePath string) (bool, error) {
	// Create a temporary file to store the search patterns
	searchPatternFile, err := createSearchPatternFile(policy.Regex, NormalizeFilename(policy.ID))
	if err != nil {
		return false, fmt.Errorf("error creating search pattern file: %w", err)
	}
	defer os.Remove(searchPatternFile)

	// Prepare the ripgrep command
	args := []string{
		"--pcre2",
		"--no-heading",
		"-o",
		"-p",
		"-i",
		"-U",
		"-f", searchPatternFile,
		filePath,
	}

	// Execute the ripgrep command
	cmd := exec.Command(rgPath, args...)
	output, err := cmd.CombinedOutput()

	// Check if ripgrep found any matches
	matchesFound := len(output) > 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 1 in ripgrep means "no matches found"
			if exitError.ExitCode() == 1 {
				matchesFound = false
				err = nil // Reset error as this is the expected outcome when no matches are found
			}
		}
		if err != nil {
			return false, fmt.Errorf("error executing ripgrep: %w", err)
		}
	}

	return matchesFound, nil
}
