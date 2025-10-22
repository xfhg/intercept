//go:build windows && amd64
// +build windows,amd64

package cmd

import "fmt"

// startSSHServer is not supported on Windows amd64 builds. This stub exists to
// satisfy references from other code paths if needed. We intentionally do NOT
// start the embedded SSH server on Windows.
func startSSHServer(policies []Policy, outputDir string) error {
	return fmt.Errorf("remote SSH server is not supported on Windows amd64 builds")
}
