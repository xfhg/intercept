//go:build linux && amd64
// +build linux,amd64

package cmd

import (
	"embed"
	"fmt"
	"os"
)

//go:embed gossh/gossh-linux-amd64
var embeddedGossh embed.FS

var gosshPath string

func prepareGosshExecutable() (string, error) {
	tempDir, err := os.MkdirTemp("", "temp_exec")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	gosshPath, err = extractExecutable(embeddedGossh, tempDir, "gossh/gossh-linux-amd64")
	if err != nil {
		return "", err
	}

	return gosshPath, nil
}