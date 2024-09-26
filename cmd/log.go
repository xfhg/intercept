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

	var details []Result
	var summary Result

	for _, result := range sarifReport.Runs[0].Results {

		policyID = result.RuleID
		resultLevel = sarifLevelToString(result.Level)
		resultLevelInt = sarifLevelToInt(result.Level)

		resultBytes, _ := json.Marshal(result)

		if result.Properties.ResultType == "summary" {

			summary = result

			if logTypeMatrixConfig.Results {
				flog.Log().Str("policy-id", policyID).Str("result-level", resultLevel).Int("sarif-level", resultLevelInt).RawJSON("summary", resultBytes).Send()
			}
			if logTypeMatrixConfig.One {
				olog.Log().Str("policy-id", policyID).Str("result-level", resultLevel).Int("sarif-level", resultLevelInt).RawJSON("summary", resultBytes).Send()
			}

		} else {

			details = append(details, result)

			if logTypeMatrixConfig.Results {
				flog.Log().Str("policy-id", policyID).Str("result-level", resultLevel).Int("sarif-level", resultLevelInt).RawJSON("detail", resultBytes).Send()
			}
			if logTypeMatrixConfig.One {
				olog.Log().Str("policy-id", policyID).Str("result-level", resultLevel).Int("sarif-level", resultLevelInt).RawJSON("detail", resultBytes).Send()
			}
		}

	}

	summaryBytes, _ := json.Marshal(summary)
	detailsBytes, _ := json.Marshal(details)

	payloadBytes, _ := json.Marshal(sarifReport.Runs[0].Results)

	if detailsBytes == nil {
		detailsBytes = []byte("[]")
	}
	if summaryBytes == nil {
		summaryBytes = []byte("{}")
	}

	if logTypeMatrixConfig.Minimal {
		mlog.Log().Str("policy-id", policyID).Bool("policy-compliant", sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant).Send()
	}

	if logTypeMatrixConfig.Policy {
		plog.Log().Bool("policy-compliant", sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant).RawJSON("summary", summaryBytes).RawJSON("results", detailsBytes).Send()
	}

	if logTypeMatrixConfig.One {
		olog.Log().Str("policy-id", policyID).Bool("policy-compliant", sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant).Send()
		olog.Log().Str("policy-id", policyID).RawJSON("policy", payloadBytes).Send()
	}

	return nil
}

func PostReportToComplianceLog(sarifReport SARIFReport) error {

	if len(sarifReport.Runs) == 0 {
		return nil
	}

	var details []Result
	var summary Result

	// Split the Results
	for _, result := range sarifReport.Runs[0].Results {

		if result.Properties.ResultType == "summary" {
			summary = result
		} else {
			details = append(details, result)
		}

	}

	payloadBytes, _ := json.Marshal(sarifReport)
	summaryBytes, _ := json.Marshal(summary)
	detailsBytes, _ := json.Marshal(details)

	if logTypeMatrixConfig.Report {
		rlog.Log().Bool("report-compliant", sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant).RawJSON("summary", summaryBytes).RawJSON("results", detailsBytes).Send()
	}
	if logTypeMatrixConfig.One {
		olog.Log().Bool("report-compliant", sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant).Str("report-status", sarifReport.Runs[0].Invocations[0].Properties.ReportStatus).Str("report-timestamp", sarifReport.Runs[0].Invocations[0].Properties.ReportTimestamp).Send()
		olog.Log().Bool("report-compliant", sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant).RawJSON("report", payloadBytes).Send()
	}
	if logTypeMatrixConfig.Minimal {
		mlog.Log().Bool("report-compliant", sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant).Str("report-status", sarifReport.Runs[0].Invocations[0].Properties.ReportStatus).Str("report-timestamp", sarifReport.Runs[0].Invocations[0].Properties.ReportTimestamp).Send()
	}

	return nil
}
