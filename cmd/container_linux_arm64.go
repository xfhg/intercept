//go:build linux && arm64 && container
// +build linux,arm64,container

package cmd

import (
	"embed"
	"fmt"
	"os"
)

//go:embed goss/goss-linux-arm64
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

	gossPath, err := extractExecutable(embeddedFiles, tempDir, "goss/goss-linux-arm64")
	if err != nil {
		return "", "", err
	}

	return rgPath, gossPath, nil
}