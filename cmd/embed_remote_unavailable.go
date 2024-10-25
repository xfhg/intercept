//go:build windows || (linux && arm && !arm64)
// +build windows linux,arm,!arm64

package cmd

import (
	"embed"
)

var embeddedGossh embed.FS

var gosshPath string

var observeRemoteHost string = "0.0.0.0"
var observeRemotePort string = "23234"

func prepareGosshExecutable() (string, error) {
	return "", nil
}

func extractGosshExecutable(_, _ string) (string, error) {
	return "", nil
}
