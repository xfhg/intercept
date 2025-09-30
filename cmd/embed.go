package cmd

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func extractExecutable(embeddedFiles embed.FS, tempDir, executableName string) (string, error) {
	executableFolder := filepath.Dir(executableName)
	err := os.MkdirAll(filepath.Join(tempDir, executableFolder), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create folder structure: %w", err)
	}

	executablePath := filepath.Join(tempDir, executableName)

	data, err := embeddedFiles.ReadFile(executableName)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded file: %w", err)
	}

	err = os.WriteFile(executablePath, data, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to write executable to temp path: %w", err)
	}

	return executablePath, nil
}

func findSystemRg() (string, error) {
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return "", fmt.Errorf("failed to find 'rg' in system PATH: %w", err)
	}
	return rgPath, nil
}