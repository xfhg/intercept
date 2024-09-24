package cmd

import (
	"encoding/json"
	"strings"
)

type LogPayload struct {
	EventType string      `json:"event_type"`
	Data      interface{} `json:"data"`
	Summary   interface{} `json:"summary,omitempty"`
	Results   interface{} `json:"results,omitempty"`
}

func containsLogType(slice []string, tag string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, tag) {
			return true
		}
	}
	return false
}

func PostResultsToComplianceLog(sarifReport SARIFReport) error {

	if len(sarifReport.Runs) == 0 {
		return nil
	}

	var policyID string
	var resultLevel string
	var resultLevelInt int

	for _, result := range sarifReport.Runs[0].Results {
		policyID = result.RuleID
		resultLevel = sarifLevelToString(result.Level)
		resultLevelInt = sarifLevelToInt(result.Level)

		if result.Properties["result-type"] == "summary" {
			resultBytes, _ := json.Marshal(result)
			clog.Log().Str("policy-id", policyID).Str("result-level", resultLevel).Int("sarif-level", resultLevelInt).RawJSON("summary", resultBytes).Send()
		} else {
			resultBytes, _ := json.Marshal(result)
			clog.Log().Str("policy-id", policyID).Str("result-level", resultLevel).Int("sarif-level", resultLevelInt).RawJSON("detail", resultBytes).Send()
		}

	}

	if sLog {
		payloadBytes, _ := json.Marshal(sarifReport.Runs[0].Results)
		clog.Log().Str("policy-id", policyID).RawJSON("policy", payloadBytes).Send()
	}
	return nil
}

func PostReportToComplianceLog(sarifReport SARIFReport) error {

	if len(sarifReport.Runs) == 0 {
		return nil
	}

	payloadBytes, _ := json.Marshal(sarifReport)

	clog.Log().Str("intercept-run-id", sarifReport.Runs[0].Invocations[0].Properties["run-id"]).Str("report-compliant", sarifReport.Runs[0].Invocations[0].Properties["report-compliant"]).Str("report-status", sarifReport.Runs[0].Invocations[0].Properties["report-status"]).Str("report-timestamp", sarifReport.Runs[0].Invocations[0].Properties["report-timestamp"]).Send()

	if sLog {
		clog.Log().Str("intercept-run-id", sarifReport.Runs[0].Invocations[0].Properties["run-id"]).Str("report-compliant", sarifReport.Runs[0].Invocations[0].Properties["report-compliant"]).RawJSON("report", payloadBytes).Send()
	}

	return nil
}
