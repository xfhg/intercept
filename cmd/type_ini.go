package cmd

// ProcessINIType delegates the processing of INI files to the generic ProcessFileType function.
func ProcessINIType(policy Policy, targetDir string, filePaths []string) error {
	return ProcessFileType(policy, targetDir, filePaths, "ini")
}