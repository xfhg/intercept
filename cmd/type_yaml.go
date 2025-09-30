package cmd

// ProcessYAMLType delegates the processing of YAML files to the generic ProcessFileType function.
func ProcessYAMLType(policy Policy, targetDir string, filePaths []string) error {
	return ProcessFileType(policy, targetDir, filePaths, "yaml")
}