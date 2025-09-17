package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

func TestExecuteAssureParallelStatusHandler(t *testing.T) {
	rgBinary, err := exec.LookPath("rg")
	if err != nil {
		t.Fatalf("rg binary not found: %v", err)
	}

	originalPolicyData := policyData
	originalOutputDir := outputDir
	policyData = &PolicyFile{}
	baseOutputDir := t.TempDir()
	outputDir = baseOutputDir
	requiredDirs := []string{"_debug", "_sarif"}
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(filepath.Join(outputDir, dir), 0o755); err != nil {
			t.Fatalf("failed to create %s dir: %v", dir, err)
		}
	}
	t.Cleanup(func() {
		policyData = originalPolicyData
		outputDir = originalOutputDir
	})

	totalFiles := parallelBatchSize + 5

	tests := []struct {
		name       string
		matchIndex int
		wantMatch  bool
	}{
		{name: "match", matchIndex: 3, wantMatch: true},
		{name: "no_match", matchIndex: -1, wantMatch: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			targetDir := filepath.Join(baseOutputDir, tc.name)
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				t.Fatalf("failed to create target dir: %v", err)
			}

			files := make([]string, 0, totalFiles)
			for i := 0; i < totalFiles; i++ {
				filePath := filepath.Join(targetDir, fmt.Sprintf("file_%s_%d.txt", tc.name, i))
				content := "this file does not contain the pattern"
				if tc.matchIndex >= 0 && i == tc.matchIndex {
					content = "this file includes the special pattern"
				}
				if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				files = append(files, filePath)
			}

			policy := Policy{
				ID:          fmt.Sprintf("policy-%s", tc.name),
				FilePattern: "*.txt",
				Regex:       []string{"special"},
			}

			var (
				mu       sync.Mutex
				statuses []bool
			)
			SetPolicyStatusHandler(func(p Policy, matched bool) {
				mu.Lock()
				defer mu.Unlock()
				statuses = append(statuses, matched)
			})
			t.Cleanup(func() {
				SetPolicyStatusHandler(nil)
			})

			if err := executeAssure(policy, rgBinary, targetDir, files); err != nil {
				t.Fatalf("executeAssure failed: %v", err)
			}

			mu.Lock()
			defer mu.Unlock()
			if len(statuses) == 0 {
				t.Fatalf("status handler was not invoked")
			}
			got := statuses[len(statuses)-1]
			if got != tc.wantMatch {
				t.Fatalf("expected match status %v, got %v", tc.wantMatch, got)
			}
		})
	}
}
