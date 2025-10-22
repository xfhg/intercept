//go:build windows && amd64
// +build windows,amd64

package cmd

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// HookStandardPayload represents the standard webhook payload structure
type HookStandardPayload struct {
	WebhookID      string      `json:"webhook-id"`
	Timestamp      string      `json:"time"`
	InterceptRunID string      `json:"intercept-run-id"`
	HostID         string      `json:"host-id"`
	Events         interface{} `json:"events"`
	EventCount     int         `json:"event-count"`
}

// NDJSONBuilder helps build NDJSON payloads
type NDJSONBuilder struct {
	buffer []string
	mutex  sync.Mutex
}

func NewNDJSONBuilder() *NDJSONBuilder {
	return &NDJSONBuilder{
		buffer: make([]string, 0),
	}
}

func (nb *NDJSONBuilder) AddLogObject(logObj interface{}) error {
	jsonBytes, err := json.Marshal(logObj)
	if err != nil {
		return fmt.Errorf("error marshaling log object: %v", err)
	}

	nb.mutex.Lock()
	nb.buffer = append(nb.buffer, string(jsonBytes))
	nb.mutex.Unlock()

	return nil
}

func (nb *NDJSONBuilder) GetNDJSON() string {
	nb.mutex.Lock()
	defer nb.mutex.Unlock()
	return strings.Join(nb.buffer, "\n")
}

func convertLogsToNDJSON(logs interface{}) (string, error) {
	reflectValue := reflect.ValueOf(logs)
	if reflectValue.Kind() != reflect.Slice {
		return "", fmt.Errorf("input must be a slice")
	}

	var ndjsonLines []string
	for i := 0; i < reflectValue.Len(); i++ {
		logEntry := reflectValue.Index(i).Interface()
		jsonBytes, err := json.Marshal(logEntry)
		if err != nil {
			return "", fmt.Errorf("error marshaling log entry: %v", err)
		}
		ndjsonLines = append(ndjsonLines, string(jsonBytes))
	}

	return strings.Join(ndjsonLines, "\n"), nil
}

// GenerateWebhookSecret generates a secure webhook secret for Windows builds
func GenerateWebhookSecret() (string, error) {
	// Generate 32 bytes of random data
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode the random bytes to base64
	secret := base64.URLEncoding.EncodeToString(randomBytes)

	// Prefix the secret with "whsec_" to match the format of the example
	return "whsec_" + secret, nil
}

func calculateWebhookSignature(payload []byte, secret string) (string, error) {
	// Split the secret and decode the base64 part
	parts := strings.SplitN(secret, "_", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid secret format")
	}
	secretBytes, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode secret: %w", err)
	}

	// Create the HMAC
	h := hmac.New(sha256.New, secretBytes)
	h.Write(payload)

	// Get the result and encode to base64
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature, nil
}

func containsEventType(slice []string, tag string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, tag) {
			return true
		}
	}
	return false
}

// PostResultsToWebhooks posts SARIF results to configured webhooks on Windows
func PostResultsToWebhooks(sarifReport SARIFReport) error {
	timestamp := time.Now().Format(time.RFC3339)

	if len(sarifReport.Runs) == 0 {
		return nil
	}
	var payload interface{}
	var payloadType string

	logMin, logRes, logPol, _, _ := processSARIF2LogStruct(sarifReport, 1, false)

	config := GetConfig()

	for _, hook := range config.Hooks {

		if containsEventType(hook.EventTypes, "minimal") {
			if containsEventType(hook.EventTypes, "log") {
				payload, _ = convertLogsToNDJSON(logMin)
				payloadType = "x-ndjson"
			} else {
				payload = HookStandardPayload{
					WebhookID:      NormalizePolicyName(hook.Name),
					Timestamp:      timestamp,
					InterceptRunID: intercept_run_id,
					HostID:         hostData,
					Events:         logMin,
					EventCount:     len(logMin),
				}
				payloadType = "json"
			}

		}
		if containsEventType(hook.EventTypes, "results") {
			if containsEventType(hook.EventTypes, "log") {
				payload, _ = convertLogsToNDJSON(logRes)
				payloadType = "x-ndjson"
			} else {

				payload = HookStandardPayload{
					WebhookID:      NormalizePolicyName(hook.Name),
					Timestamp:      timestamp,
					InterceptRunID: intercept_run_id,
					HostID:         hostData,
					Events:         logRes,
					EventCount:     len(logRes),
				}
				payloadType = "json"
			}
		}
		if containsEventType(hook.EventTypes, "policy") {

			if containsEventType(hook.EventTypes, "log") {
				payload, _ = convertLogsToNDJSON(logPol)
				payloadType = "x-ndjson"
			} else {
				payload = HookStandardPayload{
					WebhookID:      NormalizePolicyName(hook.Name),
					Timestamp:      timestamp,
					InterceptRunID: intercept_run_id,
					HostID:         hostData,
					Events:         logPol,
					EventCount:     len(logPol),
				}
				payloadType = "json"
			}
		}
		if containsEventType(hook.EventTypes, "report") {
			// same as "policy" in this context
			continue
		}

		// Check if payload is empty
		if payload == nil || payload == "" {
			log.Debug().Str("hook", hook.Name).Msg("Payload is empty, skipping webhook")
			continue
		}

		client := resty.New()
		client.SetTimeout(time.Duration(hook.TimeoutSeconds) * time.Second)
		if hook.Insecure {
			client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		}

		// Prepare the request
		req := client.R()
		req.SetHeaders(hook.Headers)

		// Set custom User-Agent
		userAgent := fmt.Sprintf("intercept/%s", buildVersion)
		req.SetHeader("User-Agent", userAgent)

		// Convert payload to []byte for signature calculation
		var payloadBytes []byte
		if payloadType == "x-ndjson" {
			payloadBytes = []byte(payload.(string))
		} else {
			payloadBytes, _ = json.Marshal(payload)
		}
		// Calculate and set the signature
		signature, _ := calculateWebhookSignature(payloadBytes, webhookSecret)
		req.SetHeader("X-Signature", signature)

		req.SetBody(payload)

		// Set content type based on payloadType
		if payloadType == "x-ndjson" {
			req.SetHeader("Content-Type", "application/x-ndjson")
		} else {
			req.SetHeader("Content-Type", "application/json")
		}

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
		// Flatten hook.EventTypes to a single string
		flatEventTypes := strings.Join(hook.EventTypes, ",")

		if err != nil {
			log.Error().Err(err).Str("hook", hook.Name).Str("event_type", flatEventTypes).Msg("Failed to post to webhook")
		} else if !resp.IsSuccess() {
			log.Error().Str("hook", hook.Name).Str("event_type", flatEventTypes).Int("status", resp.StatusCode()).Msg("Webhook request failed")
		} else {
			log.Info().Str("hook", hook.Name).Str("event_type", flatEventTypes).Str("payload_type", payloadType).Msg("Successfully posted to webhook")
		}
	}

	return nil
}

// PostReportToWebhooks posts SARIF reports to configured webhooks on Windows
func PostReportToWebhooks(sarifReport SARIFReport) error {
	timestamp := time.Now().Format(time.RFC3339)

	if len(sarifReport.Runs) == 0 {
		return nil
	}
	var payload interface{}

	var payloadType string

	_, logRes, _, logRep, _ := processSARIF2LogStruct(sarifReport, 2, false)

	config := GetConfig()

	for _, hook := range config.Hooks {

		if containsEventType(hook.EventTypes, "minimal") {

			continue
		}
		if containsEventType(hook.EventTypes, "results") {

			if containsEventType(hook.EventTypes, "log") {
				payload, _ = convertLogsToNDJSON(logRes)
				payloadType = "x-ndjson"
			} else if containsEventType(hook.EventTypes, "bulk") {

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
				payloadType = "x-ndjson"

			} else {

				payload = HookStandardPayload{
					WebhookID:      NormalizePolicyName(hook.Name),
					Timestamp:      timestamp,
					InterceptRunID: intercept_run_id,
					HostID:         hostData,
					Events:         logRes,
					EventCount:     len(logRes),
				}
				payloadType = "json"

			}
		}
		if containsEventType(hook.EventTypes, "policy") {

			continue
		}
		if containsEventType(hook.EventTypes, "report") {

			if containsEventType(hook.EventTypes, "log") {
				payload, _ = convertLogsToNDJSON(logRep)
				payloadType = "x-ndjson"
			} else {
				payload = HookStandardPayload{
					WebhookID:      NormalizePolicyName(hook.Name),
					Timestamp:      timestamp,
					InterceptRunID: intercept_run_id,
					HostID:         hostData,
					Events:         logRep,
					EventCount:     len(logRep),
				}
				payloadType = "json"
			}
		}

		// Check if payload is empty

		if payload == nil || payload == "" {
			log.Debug().Str("hook", hook.Name).Msg("Payload is empty, skipping webhook")
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

		// Set custom User-Agent
		userAgent := fmt.Sprintf("intercept/%s", buildVersion)
		req.SetHeader("User-Agent", userAgent)

		// Convert payload to []byte for signature calculation
		var payloadBytes []byte
		if payloadType == "x-ndjson" {
			payloadBytes = []byte(payload.(string))
		} else {
			payloadBytes, _ = json.Marshal(payload)
		}
		// Calculate and set the signature
		signature, _ := calculateWebhookSignature(payloadBytes, webhookSecret)
		req.SetHeader("X-Signature", signature)

		req.SetBody(payload)

		// if containsString(hook.EventTypes, "bulk") {
		// 	req.SetBody(esbulkpayload)
		// }
		if payloadType == "x-ndjson" {
			req.SetHeader("Content-Type", "application/x-ndjson")
		} else {
			req.SetHeader("Content-Type", "application/json")
		}

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

		// Flatten hook.EventTypes to a single string
		flatEventTypes := strings.Join(hook.EventTypes, ",")

		if err != nil {
			log.Error().Err(err).Str("hook", hook.Name).Str("event_type", flatEventTypes).Msg("Failed to post to webhook")
		} else if !resp.IsSuccess() {
			log.Error().Str("hook", hook.Name).Str("event_type", flatEventTypes).Int("status", resp.StatusCode()).Msg("Webhook request failed")
		} else {
			log.Info().Str("hook", hook.Name).Str("event_type", flatEventTypes).Str("payload_type", payloadType).Msg("Successfully posted to webhook")
		}
	}

	return nil
}
