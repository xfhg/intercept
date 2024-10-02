package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Consistent Logging and Integration Data with (hook.go)

type LogMinimal struct {
	PolicyCompliant bool   `json:"policy-compliant"`
	PolicyID        string `json:"policy-id"`
}

type LogResults struct {
	PolicyCompliant bool   `json:"policy-compliant"`
	PolicyID        string `json:"policy-id"`
	SarifLevelInt   int    `json:"sarif-level-int"`
	SarifLevel      string `json:"sarif-level"`

	ResultType string `json:"result-type"`
	Result     Result `json:"result"`
}
type LogPolicy struct {
	PolicyCompliant bool   `json:"policy-compliant"`
	PolicyID        string `json:"policy-id"`

	Summary Result   `json:"summary,omitempty"`
	Results []Result `json:"results,omitempty"`

	PolicyProperties interface{} `json:"policy-properties,omitempty"`
}

type LogReport struct {
	ReportCompliant bool   `json:"report-compliant"`
	ReportID        string `json:"report-id"`

	Summary []Result `json:"summary,omitempty"`
	Results []Result `json:"results,omitempty"`

	ReportProperties interface{} `json:"report-properties,omitempty"`
}

func processSARIF2LogStruct(sarifReport SARIFReport, payloadType int) ([]LogMinimal, []LogResults, []LogPolicy, []LogReport, error) {

	if len(sarifReport.Runs) == 0 {
		return nil, nil, nil, nil, nil
	}

	switch payloadType {
	case 1:
		// Individual Report
		var individualLogMinimal LogMinimal
		var iLogMin []LogMinimal
		var individualLogResults LogResults
		var iLogRes []LogResults
		var individualLogPolicy LogPolicy
		var iLogPol []LogPolicy

		var policyID string
		var resultLevel string
		var resultLevelInt int
		var policyCompliant bool

		var detailsResults []Result
		var summaryResults []Result

		policyCompliant = sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant

		for _, result := range sarifReport.Runs[0].Results {

			policyID = result.RuleID
			resultLevel = sarifLevelToString(result.Level)
			resultLevelInt = sarifLevelToInt(result.Level)

			individualLogResults.PolicyCompliant = policyCompliant
			individualLogResults.PolicyID = policyID
			individualLogResults.Result = result
			individualLogResults.ResultType = result.Properties.ResultType
			individualLogResults.SarifLevelInt = resultLevelInt
			individualLogResults.SarifLevel = resultLevel

			iLogRes = append(iLogRes, individualLogResults)

			if result.Properties.ResultType == "summary" {
				summaryResults = append(summaryResults, result)
			} else {
				detailsResults = append(detailsResults, result)
			}

		}

		individualLogPolicy.PolicyCompliant = policyCompliant
		individualLogPolicy.PolicyID = policyID
		individualLogPolicy.PolicyProperties = sarifReport.Runs[0].Invocations[0].Properties
		individualLogPolicy.Results = detailsResults
		if len(summaryResults) > 0 {
			individualLogPolicy.Summary = summaryResults[0]
		} else {
			individualLogPolicy.Summary = Result{
				RuleID: policyID,
				Level:  "note",
				Message: Message{
					Text: "No summary results for this policy",
				},
			}
		}

		individualLogMinimal.PolicyCompliant = policyCompliant
		individualLogMinimal.PolicyID = policyID

		iLogPol = append(iLogPol, individualLogPolicy)
		iLogMin = append(iLogMin, individualLogMinimal)

		return iLogMin, iLogRes, iLogPol, nil, nil

	case 2:
		// Merged Report
		var mergedLogReport LogReport
		var mLogRep []LogReport

		var mergedLogResults LogResults
		var mLogRes []LogResults

		var reportCompliant bool

		var policyID string
		var resultLevel string
		var resultLevelInt int
		var policyCompliant bool

		var detailsResults []Result
		var summaryResults []Result

		reportCompliant = sarifReport.Runs[0].Invocations[0].Properties.ReportCompliant

		originalTime, _ := time.Parse(time.RFC3339, sarifReport.Runs[0].Invocations[0].Properties.ReportTimestamp)

		for _, result := range sarifReport.Runs[0].Results {

			policyID = result.RuleID
			resultLevel = sarifLevelToString(result.Level)
			resultLevelInt = sarifLevelToInt(result.Level)

			mergedLogResults.PolicyCompliant = policyCompliant
			mergedLogResults.PolicyID = policyID
			mergedLogResults.Result = result
			mergedLogResults.ResultType = result.Properties.ResultType
			mergedLogResults.SarifLevelInt = resultLevelInt
			mergedLogResults.SarifLevel = resultLevel

			mLogRes = append(mLogRes, mergedLogResults)

			if result.Properties.ResultType == "summary" {
				summaryResults = append(summaryResults, result)
			} else {
				detailsResults = append(detailsResults, result)
			}

		}

		mergedLogReport.ReportCompliant = reportCompliant
		mergedLogReport.ReportID = fmt.Sprintf("%s_%s", originalTime.UTC().Format("20060102T150405Z"), sarifReport.Runs[0].Invocations[0].Properties.RunId)
		mergedLogReport.ReportProperties = sarifReport.Runs[0].Invocations[0].Properties
		mergedLogReport.Summary = summaryResults
		mergedLogReport.Results = detailsResults

		mLogRep = append(mLogRep, mergedLogReport)

		return nil, mLogRes, nil, mLogRep, nil

	}

	return nil, nil, nil, nil, nil

}

func PostResultsToComplianceLog(sarifReport SARIFReport) error {

	if len(sarifReport.Runs) == 0 {
		return nil
	}

	logMin, logRes, logPol, _, _ := processSARIF2LogStruct(sarifReport, 1)

	if logTypeMatrixConfig.Minimal {
		for _, log := range logMin {
			mlog.Log().
				Str("policy-id", log.PolicyID).
				Bool("policy-compliant", log.PolicyCompliant).
				Send()
		}
	}
	if logTypeMatrixConfig.Results {
		for _, log := range logRes {
			rBytes, _ := json.Marshal(log.Result)
			if rBytes == nil {
				rBytes = []byte("{}")
			}
			flog.Log().
				Str("policy-id", log.PolicyID).
				Bool("policy-compliant", log.PolicyCompliant).
				Str("sarif-level", log.SarifLevel).
				Int("sarif-level-int", log.SarifLevelInt).
				Str("result-type", log.ResultType).
				RawJSON("result", rBytes).
				Send()
		}
	}
	if logTypeMatrixConfig.Policy {
		for _, log := range logPol {
			rsBytes, _ := json.Marshal(log.Summary)
			rrBytes, _ := json.Marshal(log.Results)
			if rsBytes == nil {
				rsBytes = []byte("[]")
			}
			if rrBytes == nil {
				rrBytes = []byte("[]")
			}
			plog.Log().
				Str("policy-id", log.PolicyID).
				Bool("policy-compliant", log.PolicyCompliant).
				RawJSON("summary", rsBytes).
				RawJSON("results", rrBytes).
				Send()
		}
	}

	return nil
}

func PostReportToComplianceLog(sarifReport SARIFReport) error {

	if len(sarifReport.Runs) == 0 {
		return nil
	}

	if logTypeMatrixConfig.Report {
		_, _, _, logRep, _ := processSARIF2LogStruct(sarifReport, 2)

		for _, log := range logRep {
			rsBytes, _ := json.Marshal(log.Summary)
			rrBytes, _ := json.Marshal(log.Results)
			rpBytes, _ := json.Marshal(log.ReportProperties)
			if rsBytes == nil {
				rsBytes = []byte("[]")
			}
			if rrBytes == nil {
				rrBytes = []byte("[]")
			}
			rlog.Log().
				Str("report-id", log.ReportID).
				Bool("report-compliant", log.ReportCompliant).
				RawJSON("summary", rsBytes).
				RawJSON("results", rrBytes).
				RawJSON("report-properties", rpBytes).
				Send()
		}
	}
	return nil
}

func containsLogType(slice []string, tag string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, tag) {
			return true
		}
	}
	return false
}
