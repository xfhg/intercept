//go:build linux && arm && !container
// +build linux,arm,!container

package cmd

import (
	"embed"
	"fmt"
	"os"
)

//go:embed rg/rg-linux-arm goss/goss-linux-arm
var embeddedFiles embed.FS

func prepareEmbeddedExecutables() (string, string, error) {
	tempDir, err := os.MkdirTemp("", "temp_exec")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	rgPath, err = extractExecutable(embeddedFiles, tempDir, "rg/rg-linux-arm")
	if err != nil {
		return "", "", err
	}

	gossPath, err = extractExecutable(embeddedFiles, tempDir, "goss/goss-linux-arm")
	if err != nil {
		return "", "", err
	}

	return rgPath, gossPath, nil
}