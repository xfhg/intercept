//go:build windows || (linux && arm && !arm64)
// +build windows linux,arm,!arm64

package cmd

import (
	"embed"
	"fmt"
)

var embeddedGossh embed.FS

var gosshPath string

func prepareGosshExecutable() (string, error) {
	return "", fmt.Errorf("gossh executable not available on this platform")
}

func extractGosshExecutable(_, _ string) (string, error) {
	return "", fmt.Errorf("gossh executable not available on this platform")
}
