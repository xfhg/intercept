package cmd

// ProcessTOMLType delegates the processing of TOML files to the generic ProcessFileType function.
func ProcessTOMLType(policy Policy, targetDir string, filePaths []string) error {
	return ProcessFileType(policy, targetDir, filePaths, "toml")
}