//go:build darwin && arm64
// +build darwin,arm64

package cmd

import (
	"embed"
	"fmt"
	"os"
)

//go:embed rg/rg-darwin-arm64 goss/goss-darwin-arm64
var embeddedFiles embed.FS

func prepareEmbeddedExecutables() (string, string, error) {
	tempDir, err := os.MkdirTemp("", "temp_exec")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	rgPath, err = extractExecutable(embeddedFiles, tempDir, "rg/rg-darwin-arm64")
	if err != nil {
		return "", "", err
	}

	gossPath, err = extractExecutable(embeddedFiles, tempDir, "goss/goss-darwin-arm64")
	if err != nil {
		return "", "", err
	}

	return rgPath, gossPath, nil
}