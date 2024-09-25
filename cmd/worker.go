package cmd

import (
	"fmt"

	"github.com/gookit/event"
)

func preparePolicyPaths(policy Policy, allFileInfos []FileInfo) []string {

	_, filePaths := filterFiles(policy, allFileInfos)
	return filePaths
}

// processPolicyInWorker handles policy processing based on the policy type
func processPolicyInWorker(e event.Event, policyType string) error {
	policy, ok := e.Get("policy").(Policy)
	if !ok {
		return fmt.Errorf("invalid policy data for %s", policyType)
	}

	// this data only needed for non-runtime non-api policies
	targetDir, _ := e.Get("targetDir").(string)
	filePaths, _ := e.Get("filePaths").([]string)

	if policy.ID == "" {
		log.Error().Msg("Error: policy ID is empty, can't proceed with policy auditing")
		return fmt.Errorf("policy ID is empty, can't proceed with policy auditing")
	}

	if policy.Type == "runtime" || policy.Type == "api" {
		log.Debug().Str("policy", policy.ID).Str("type", policy.Type).Msgf("Working")
	} else {
		log.Debug().Str("policy", policy.ID).Str("type", policy.Type).Msgf("Working [%s] [%s]", targetDir, filePaths)
	}

	switch policyType {
	case "scan":
		return ProcessScanType(policy, rgPath, targetDir, filePaths)
	case "assure":
		return ProcessAssureType(policy, rgPath, targetDir, filePaths)
	case "runtime":
		return ProcessRuntimeType(policy, gossPath, targetDir, filePaths, true)
	case "api":
		return ProcessAPIType(policy, rgPath)
	case "yml":
		if policy.Schema.Patch {
			return processGenericType(policy, filePaths, "yaml")
		}
		return ProcessYAMLType(policy, targetDir, filePaths)
	case "toml":
		if policy.Schema.Patch {
			return processGenericType(policy, filePaths, "toml")
		}
		return ProcessTOMLType(policy, targetDir, filePaths)
	case "json":
		if policy.Schema.Patch {
			return ProcessJSONTypeWithPatch(policy, targetDir, filePaths)
		}
		return ProcessJSONType(policy, targetDir, filePaths)
	case "ini":
		if policy.Schema.Patch {
			return processGenericType(policy, filePaths, "ini")
		}
		return ProcessINIType(policy, targetDir, filePaths)
	case "rego":
		return ProcessRegoType(policy, targetDir, filePaths)
	default:
		return fmt.Errorf("unsupported policy type: %s", policyType)
	}

}
