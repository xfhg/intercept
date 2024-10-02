//go:build linux && amd64 && container
// +build linux,amd64,container

package cmd

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed goss/goss-linux-amd64
var embeddedFiles embed.FS

func prepareEmbeddedExecutables() (string, string, error) {
	tempDir, err := os.MkdirTemp("", "temp_exec")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	rgPath, err := findSystemRg()
	if err != nil {
		return "", "", err
	}

	gossPath, err := extractExecutable(tempDir, "goss/goss-linux-amd64")
	if err != nil {
		return "", "", err
	}

	return rgPath, gossPath, nil
}

func findSystemRg() (string, error) {
	rgPath, err := exec.LookPath("rg")
	if err != nil {
		return "", fmt.Errorf("failed to find 'rg' in system PATH: %w", err)
	}
	return rgPath, nil
}

func extractExecutable(tempDir, executableName string) (string, error) {
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
