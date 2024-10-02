//go:build linux && arm64 && !container
// +build linux,arm64,!container

package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed rg/rg-linux-arm64 goss/goss-linux-arm64
var embeddedFiles embed.FS

func prepareEmbeddedExecutables() (string, string, error) {
	tempDir, err := os.MkdirTemp("", "temp_exec")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	rgPath, err = extractExecutable(tempDir, "rg/rg-linux-arm64")
	if err != nil {
		return "", "", err
	}

	gossPath, err = extractExecutable(tempDir, "goss/goss-linux-arm64")
	if err != nil {
		return "", "", err
	}

	return rgPath, gossPath, nil
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
