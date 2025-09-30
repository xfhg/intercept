package cmd

// ProcessJSONType delegates the processing of JSON files to the generic ProcessFileType function.
func ProcessJSONType(policy Policy, targetDir string, filePaths []string) error {
	return ProcessFileType(policy, targetDir, filePaths, "json")
}

// ProcessJSONTypeWithPatch delegates the processing and patching of JSON files to the generic ProcessFileTypeWithPatch function.
func ProcessJSONTypeWithPatch(policy Policy, targetDir string, filePaths []string) error {
	return ProcessFileTypeWithPatch(policy, targetDir, filePaths, "json")
}