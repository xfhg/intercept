//go:build linux && amd64 && !container
// +build linux,amd64,!container

package cmd

import (
	"embed"
	"fmt"
	"os"
)

//go:embed rg/rg-linux-amd64 goss/goss-linux-amd64
var embeddedFiles embed.FS

func prepareEmbeddedExecutables() (string, string, error) {
	tempDir, err := os.MkdirTemp("", "temp_exec")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	rgPath, err = extractExecutable(embeddedFiles, tempDir, "rg/rg-linux-amd64")
	if err != nil {
		return "", "", err
	}

	gossPath, err = extractExecutable(embeddedFiles, tempDir, "goss/goss-linux-amd64")
	if err != nil {
		return "", "", err
	}

	return rgPath, gossPath, nil
}