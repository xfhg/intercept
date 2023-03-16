//go:build linux
// +build linux

package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed rg/rgl
var embeddedFile embed.FS

func prepareEmbeddedExecutable() (string, error) {
	tempDir, err := os.MkdirTemp("", "temp_exec")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	// defer os.RemoveAll(tempDir)

	executableName := "rg/rgl"

	executableFolder := filepath.Dir(executableName)

	err = os.MkdirAll(filepath.Join(tempDir, executableFolder), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create folder structure: %w", err)
	}

	executablePath := filepath.Join(tempDir, executableName)

	data, err := embeddedFile.ReadFile(executableName)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded file: %w", err)
	}

	err = os.WriteFile(executablePath, data, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to write executable to temp path: %w", err)
	}

	return executablePath, nil
}
