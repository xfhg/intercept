package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func RunCommandWithArgs(binaryPath string, argsStr string) error {
	// Split the arguments string into a slice
	args := strings.Fields(argsStr)

	// Prepare the command
	cmd := exec.Command(binaryPath, args...)

	// Prepare buffers to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command execution failed: %v", err)
	}

	// Save stdout to a file
	if err := os.WriteFile("stdout.txt", stdout.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write stdout to file: %v", err)
	}

	// Save stderr to a file
	if err := os.WriteFile("stderr.txt", stderr.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write stderr to file: %v", err)
	}

	return nil
}
