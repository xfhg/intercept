//go:build linux && amd64
// +build linux,amd64

package cmd

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed gossh/gossh-linux-amd64
var embeddedGossh embed.FS

var gosshPath string

func prepareGosshExecutable() (string, error) {
	tempDir, err := os.MkdirTemp("", "temp_exec")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	gosshPath, err = extractGosshExecutable(tempDir, "gossh/gossh-linux-amd64")
	if err != nil {
		return "", err
	}

	return gosshPath, nil
}

func extractGosshExecutable(tempDir, executableName string) (string, error) {
	executableFolder := filepath.Dir(executableName)
	err := os.MkdirAll(filepath.Join(tempDir, executableFolder), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create folder structure: %w", err)
	}

	executablePath := filepath.Join(tempDir, executableName)

	data, err := embeddedGossh.ReadFile(executableName)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded gossh: %w", err)
	}

	err = os.WriteFile(executablePath, data, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to write gossh to temp path: %w", err)
	}

	return executablePath, nil
}
