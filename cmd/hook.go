//go:build !windows
// +build !windows

package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// type bulkWebhookPayload struct {
// 	Index string      `json:"index"`
// 	Data  interface{} `json:"data"`
// }

type HookStandardPayload struct {
	WebhookID      string      `json:"webhook-id"`
	Timestamp      string      `json:"time"`
	InterceptRunID string      `json:"intercept-run-id"`
	HostID         string      `json:"host-id"`
	Events         interface{} `json:"events"`
	EventCount     int         `json:"event-count"`
}

type DatalakePayload struct {
	WebhookID       string      `json:"webhook-id"`
	Timestamp       string      `json:"time"`
	InterceptRunID  string      `json:"intercept-run-id"`
	HostID          string      `json:"host-id"`
	PolicyCompliant bool        `json:"policy-compliant"`
	Summary         interface{} `json:"summary"`
	Results         interface{} `json:"results"`
}

// event types same as log types : minimal, results, policy, report
// custom modifiers : policy+bulk, report+bulk,

func PostResultsToWebhooks(sarifReport SARIFReport) error {

	timestamp := time.Now().Format(time.RFC3339)

	if len(sarifReport.Runs) == 0 {
		return nil
	}
	var payload interface{}

	logMin, logRes, logPol, _, _ := processSARIF2LogStruct(sarifReport, 1)

	config := GetConfig()

	for _, hook := range config.Hooks {

		if containsEventType(hook.EventTypes, "minimal") {

			payload = HookStandardPayload{
				WebhookID:      NormalizePolicyName(hook.Name),
				Timestamp:      timestamp,
				InterceptRunID: intercept_run_id,
				HostID:         hostData,
				Events:         logMin,
				EventCount:     len(logMin),
			}

		}
		if containsEventType(hook.EventTypes, "results") {

			payload = HookStandardPayload{
				WebhookID:      NormalizePolicyName(hook.Name),
				Timestamp:      timestamp,
				InterceptRunID: intercept_run_id,
				HostID:         hostData,
				Events:         logRes,
				EventCount:     len(logRes),
			}

		}
		if containsEventType(hook.EventTypes, "policy") {
			if containsEventType(hook.EventTypes, "datalake") {
				if len(logPol) > 0 {
					payload = DatalakePayload{
						WebhookID:       NormalizePolicyName(hook.Name),
						Timestamp:       timestamp,
						InterceptRunID:  intercept_run_id,
						HostID:          hostData,
						PolicyCompliant: logPol[0].PolicyCompliant,
						Summary:         logPol[0].Summary,
						Results:         logPol[0].Results,
					}
				}
			} else {
				payload = HookStandardPayload{
					WebhookID:      NormalizePolicyName(hook.Name),
					Timestamp:      timestamp,
					InterceptRunID: intercept_run_id,
					HostID:         hostData,
					Events:         logPol,
					EventCount:     len(logPol),
				}
			}
		}
		if containsEventType(hook.EventTypes, "report") {
			// same as "policy" in this context
			continue
		}

		// ----------------------------------------------
		// ----------------------------------------------
		// ----------------------------------------------

		// if !containsString(hook.EventTypes, "policy") {
		// 	continue
		// }

		// log.Info().Str("hook_name", hook.Name).Str("endpoint", hook.Endpoint).Msg("Webhooking")

		// var payload WebhookPayload
		// var esbulkpayload string

		// if containsString(hook.EventTypes, "policy") {

		// 	if len(sarifReport.Runs) == 0 {
		// 		log.Warn().Str("hook", hook.Name).Msg("SARIF report contains no runs, skipping")
		// 		continue
		// 	}

		// 	// If the event type is "results", only send the Results array
		// 	payload = WebhookPayload{
		// 		EventType: "policy",
		// 		Data:      sarifReport.Runs[0].Results,
		// 	}
		// }

		// // ----------------------------------------------
		// // ---------------------------------------------- Split Results
		// // ---------------------------------------------- Rewrites the payload with split data

		// if containsString(hook.EventTypes, "split") {

		// 	if len(sarifReport.Runs) == 0 {
		// 		log.Warn().Str("hook", hook.Name).Msg("SARIF report contains no runs, skipping")
		// 		continue
		// 	}

		// 	var details []Result
		// 	var summary Result

		// 	// Split the Results
		// 	for _, result := range sarifReport.Runs[0].Results {

		// 		if result.Properties.ResultType == "summary" {
		// 			summary = result
		// 		} else {
		// 			details = append(details, result)
		// 		}

		// 	}

		// 	// If the event type is "poc", we split the array
		// 	payload = WebhookPayload{
		// 		EventType: "policy",
		// 		Summary:   summary,
		// 		Results:   details,
		// 	}
		// }

		// // ----------------------------------------------
		// // ---------------------------------------------- Split Results END
		// // ----------------------------------------------

		// // ----------------------------------------------
		// // ---------------------------------------------- ES Bulk
		// // ---------------------------------------------- adds { "index": { "_index": "ng" } } between results
		// if containsString(hook.EventTypes, "bulk") {

		// 	if len(sarifReport.Runs) == 0 {
		// 		log.Warn().Str("hook", hook.Name).Msg("SARIF report contains no runs, skipping")
		// 		continue
		// 	}

		// 	// Create a buffer to hold the bulk payload
		// 	var bulkPayload bytes.Buffer

		// 	// Function to write an index action
		// 	writeIndexAction := func() error {
		// 		indexAction := map[string]interface{}{
		// 			"index": map[string]interface{}{
		// 				"_index": observeConfig.Flags.Index,
		// 			},
		// 		}
		// 		indexActionBytes, err := json.Marshal(indexAction)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		bulkPayload.Write(indexActionBytes)
		// 		bulkPayload.WriteByte('\n')
		// 		return nil
		// 	}

		// 	for _, result := range sarifReport.Runs[0].Results {
		// 		// Write the index action
		// 		if err := writeIndexAction(); err != nil {
		// 			log.Printf("Error marshalling index action: %v", err)
		// 			continue
		// 		}

		// 		// Write the result JSON
		// 		resultBytes, err := json.Marshal(result)
		// 		if err != nil {
		// 			log.Printf("Error marshalling result: %v", err)
		// 			continue
		// 		}
		// 		bulkPayload.Write(resultBytes)
		// 		bulkPayload.WriteByte('\n')
		// 	}

		// 	esbulkpayload = bulkPayload.String()

		// }

		// if containsString(hook.EventTypes, "bulk") {
		// 	req.SetBody(esbulkpayload)
		// }

		// ----------------------------------------------
		// ---------------------------------------------- ES Bulk Polcicies END
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

func PostReportToWebhooks(sarifReport SARIFReport) error {

	timestamp := time.Now().Format(time.RFC3339)

	if len(sarifReport.Runs) == 0 {
		return nil
	}
	var payload interface{}

	_, logRes, _, logRep, _ := processSARIF2LogStruct(sarifReport, 1)

	config := GetConfig()

	for _, hook := range config.Hooks {

		if containsEventType(hook.EventTypes, "minimal") {

			continue
		}
		if containsEventType(hook.EventTypes, "results") {
			if containsEventType(hook.EventTypes, "bulk") {

				// Create a buffer to hold the bulk payload
				var bulkPayload bytes.Buffer

				// Function to write an index action
				writeIndexAction := func() error {
					indexAction := map[string]interface{}{
						"index": map[string]interface{}{
							"_index": observeConfig.Flags.Index,
						},
					}
					indexActionBytes, err := json.Marshal(indexAction)
					if err != nil {
						return err
					}
					bulkPayload.Write(indexActionBytes)
					bulkPayload.WriteByte('\n')
					return nil
				}

				for _, result := range sarifReport.Runs[0].Results {
					// Write the index action
					if err := writeIndexAction(); err != nil {
						log.Printf("Error marshalling index action: %v", err)
						continue
					}

					// Write the result JSON
					resultBytes, err := json.Marshal(result)
					if err != nil {
						log.Printf("Error marshalling result: %v", err)
						continue
					}
					bulkPayload.Write(resultBytes)
					bulkPayload.WriteByte('\n')
				}

				payload = bulkPayload.String()

			} else {

				payload = HookStandardPayload{
					WebhookID:      NormalizePolicyName(hook.Name),
					Timestamp:      timestamp,
					InterceptRunID: intercept_run_id,
					HostID:         hostData,
					Events:         logRes,
					EventCount:     len(logRes),
				}

			}
		}
		if containsEventType(hook.EventTypes, "policy") {

			continue
		}
		if containsEventType(hook.EventTypes, "report") {
			payload = HookStandardPayload{
				WebhookID:      NormalizePolicyName(hook.Name),
				Timestamp:      timestamp,
				InterceptRunID: intercept_run_id,
				HostID:         hostData,
				Events:         logRep,
				EventCount:     len(logRep),
			}
		}

		// if !containsString(hook.EventTypes, "report") && !containsString(hook.EventTypes, "results") {
		// 	continue
		// }

		// log.Info().Str("hook_name", hook.Name).Str("endpoint", hook.Endpoint).Msg("Webhooking")

		// var payload WebhookPayload
		// var esbulkpayload string

		// if containsString(hook.EventTypes, "results") {

		// 	if len(sarifReport.Runs) == 0 {
		// 		log.Warn().Str("hook", hook.Name).Msg("SARIF report contains no runs, skipping results webhook")
		// 		continue
		// 	}

		// 	// If the event type is "results", only send the Results array
		// 	payload = WebhookPayload{
		// 		EventType: "results",
		// 		Data:      sarifReport.Runs[0].Results,
		// 	}

		// 	// ----------------------------------------------
		// 	// ---------------------------------------------- ES Bulk
		// 	// ---------------------------------------------- adds { "index": { "_index": "ng" } } between results

		// 	if containsString(hook.EventTypes, "bulk") {
		// 		// Create a buffer to hold the bulk payload
		// 		var bulkPayload bytes.Buffer

		// 		// Function to write an index action
		// 		writeIndexAction := func() error {
		// 			indexAction := map[string]interface{}{
		// 				"index": map[string]interface{}{
		// 					"_index": observeConfig.Flags.Index,
		// 				},
		// 			}
		// 			indexActionBytes, err := json.Marshal(indexAction)
		// 			if err != nil {
		// 				return err
		// 			}
		// 			bulkPayload.Write(indexActionBytes)
		// 			bulkPayload.WriteByte('\n')
		// 			return nil
		// 		}

		// 		for _, result := range sarifReport.Runs[0].Results {
		// 			// Write the index action
		// 			if err := writeIndexAction(); err != nil {
		// 				log.Printf("Error marshalling index action: %v", err)
		// 				continue
		// 			}

		// 			// Write the result JSON
		// 			resultBytes, err := json.Marshal(result)
		// 			if err != nil {
		// 				log.Printf("Error marshalling result: %v", err)
		// 				continue
		// 			}
		// 			bulkPayload.Write(resultBytes)
		// 			bulkPayload.WriteByte('\n')
		// 		}

		// 		esbulkpayload = bulkPayload.String()
		// 	}

		// } else if containsString(hook.EventTypes, "report") {
		// 	// For "report" event type, send the full SARIF report
		// 	payload = WebhookPayload{
		// 		EventType: "report",
		// 		Data:      sarifReport,
		// 	}
		// }

		// Check if payload is empty
		if payload == nil {
			log.Warn().Str("hook", hook.Name).Msg("Payload is empty, skipping webhook")
			continue
		}

		client := resty.New()
		client.SetTimeout(time.Duration(hook.TimeoutSeconds) * time.Second)
		if hook.Insecure {
			client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		}
		client.SetDebug(debugOutput)

		// Prepare the request
		req := client.R()
		req.SetHeaders(hook.Headers)
		req.SetBody(payload)

		// if containsString(hook.EventTypes, "bulk") {
		// 	req.SetBody(esbulkpayload)
		// }

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

func containsEventType(slice []string, tag string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, tag) {
			return true
		}
	}
	return false
}
