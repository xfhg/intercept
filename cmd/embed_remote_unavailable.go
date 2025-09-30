//go:build windows || (linux && arm && !arm64)
// +build windows linux,arm,!arm64

package cmd

import (
	"embed"
)

var embeddedGossh embed.FS

var gosshPath string

func prepareGosshExecutable() (string, error) {
	return "", nil
}