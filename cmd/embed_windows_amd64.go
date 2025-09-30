//go:build windows && amd64
// +build windows,amd64

package cmd

import (
	"embed"
	"fmt"
	"os"
)

//go:embed rg/rg-windows-amd64.exe goss/goss-windows-amd64.exe
var embeddedFiles embed.FS

func prepareEmbeddedExecutables() (string, string, error) {
	tempDir, err := os.MkdirTemp("", "temp_exec")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	rgPath, err = extractExecutable(embeddedFiles, tempDir, "rg/rg-windows-amd64.exe")
	if err != nil {
		return "", "", err
	}

	gossPath, err = extractExecutable(embeddedFiles, tempDir, "goss/goss-windows-amd64.exe")
	if err != nil {
		return "", "", err
	}

	return rgPath, gossPath, nil
}