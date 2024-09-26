// hook.go

package cmd

import (
	"crypto/tls"
	"time"

	"github.com/go-resty/resty/v2"
)

type WebhookPayload struct {
	EventType string      `json:"event_type"`
	Data      interface{} `json:"data"`
	Summary   interface{} `json:"summary,omitempty"`
	Results   interface{} `json:"results,omitempty"`
	Timestamp string      `json:"timestamp,omitempty"`
}

func PostReportToWebhooks(sarifReport SARIFReport) error {
	config := GetConfig()

	for _, hook := range config.Hooks {
		if !containsString(hook.EventTypes, "report") && !containsString(hook.EventTypes, "results") {
			continue
		}

		log.Info().Str("hook_name", hook.Name).Str("endpoint", hook.Endpoint).Msg("Webhooking")

		var payload WebhookPayload
		if containsString(hook.EventTypes, "results") {

			if len(sarifReport.Runs) == 0 {
				log.Warn().Str("hook", hook.Name).Msg("SARIF report contains no runs, skipping results webhook")
				continue
			}

			// If the event type is "results", only send the Results array
			payload = WebhookPayload{
				EventType: "results",
				Data:      sarifReport.Runs[0].Results,
			}
		} else {
			// For "report" event type, send the full SARIF report
			payload = WebhookPayload{
				EventType: "report",
				Data:      sarifReport,
			}
		}

		client := resty.New()
		client.SetTimeout(time.Duration(hook.TimeoutSeconds) * time.Second)
		if hook.Insecure {
			client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		}

		// Prepare the request
		req := client.R()
		req.SetHeaders(hook.Headers)
		req.SetBody(payload)

		// Apply authentication
		if err := applyAuth(req, hook.Auth); err != nil {
			log.Error().Err(err).Str("hook", hook.Name).Msg("Failed to apply authentication")
			continue
		}

		// Send the request
		var resp *resty.Response
		var err error
		for attempt := 0; attempt <= hook.RetryAttempts; attempt++ {
			resp, err = req.Execute(hook.Method, hook.Endpoint)
			if err == nil && resp.IsSuccess() {
				break
			}
			if attempt < hook.RetryAttempts {
				retryDelay, _ := time.ParseDuration(hook.RetryDelay)
				time.Sleep(retryDelay)
			}
		}

		if err != nil {
			log.Error().Err(err).Str("hook", hook.Name).Msg("Failed to post to webhook")
		} else if !resp.IsSuccess() {
			log.Error().Str("hook", hook.Name).Int("status", resp.StatusCode()).Msg("Webhook request failed")
		} else {
			log.Info().Str("hook", hook.Name).Msg("Successfully posted to webhook")
		}
	}

	return nil
}

func PostResultsToWebhooks(sarifReport SARIFReport) error {
	config := GetConfig()

	for _, hook := range config.Hooks {
		if !containsString(hook.EventTypes, "policy") {
			continue
		}

		log.Info().Str("hook_name", hook.Name).Str("endpoint", hook.Endpoint).Msg("Webhooking")

		var payload WebhookPayload
		if containsString(hook.EventTypes, "policy") {

			if len(sarifReport.Runs) == 0 {
				log.Warn().Str("hook", hook.Name).Msg("SARIF report contains no runs, skipping")
				continue
			}

			// If the event type is "results", only send the Results array
			payload = WebhookPayload{
				EventType: "policy",
				Data:      sarifReport.Runs[0].Results,
			}
		}

		// ----------------------------------------------
		// ---------------------------------------------- POC Polcicies
		// ---------------------------------------------- Rewrites the payload with split data

		if containsString(hook.EventTypes, "poc") {

			if len(sarifReport.Runs) == 0 {
				log.Warn().Str("hook", hook.Name).Msg("SARIF report contains no runs, skipping")
				continue
			}

			var details []Result
			var summary Result

			// Split the Results
			for _, result := range sarifReport.Runs[0].Results {

				if result.Properties["result-type"] == "summary" {
					summary = result
				} else {
					details = append(details, result)
				}

			}

			// If the event type is "poc", we split the array
			payload = WebhookPayload{
				EventType: "policy",
				Data:      sarifReport.Runs[0].Results,
				Summary:   summary,
				Results:   details,
			}
		}

		// ----------------------------------------------
		// ---------------------------------------------- POC Polcicies END
		// ----------------------------------------------

		client := resty.New()
		client.SetTimeout(time.Duration(hook.TimeoutSeconds) * time.Second)
		if hook.Insecure {
			client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		}

		// Prepare the request
		req := client.R()
		req.SetHeaders(hook.Headers)
		req.SetBody(payload)

		// Apply authentication
		if err := applyAuth(req, hook.Auth); err != nil {
			log.Error().Err(err).Str("hook", hook.Name).Msg("Failed to apply authentication")
			continue
		}

		// Send the request
		var resp *resty.Response
		var err error
		for attempt := 0; attempt <= hook.RetryAttempts; attempt++ {
			resp, err = req.Execute(hook.Method, hook.Endpoint)
			if err == nil && resp.IsSuccess() {
				break
			}
			if attempt < hook.RetryAttempts {
				retryDelay, _ := time.ParseDuration(hook.RetryDelay)
				time.Sleep(retryDelay)
			}
		}

		if err != nil {
			log.Error().Err(err).Str("hook", hook.Name).Msg("Failed to post to webhook")
		} else if !resp.IsSuccess() {
			log.Error().Str("hook", hook.Name).Int("status", resp.StatusCode()).Msg("Webhook request failed")
		} else {
			log.Info().Str("hook", hook.Name).Msg("Successfully posted to webhook")
		}
	}

	return nil
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
