package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"sync"
)

const parallelBatchSize = 25

type ripgrepOutputConsumer func([]byte) error

func runParallelRipgrep(rgPath string, baseArgs []string, files []string, consume ripgrepOutputConsumer) (bool, error) {
	if len(files) == 0 {
		return false, nil
	}

	var (
		matchesFound bool
		matchesMu    sync.Mutex
		wg           sync.WaitGroup
		errChan      = make(chan error, (len(files)+parallelBatchSize-1)/parallelBatchSize)
	)

	for start := 0; start < len(files); start += parallelBatchSize {
		end := start + parallelBatchSize
		if end > len(files) {
			end = len(files)
		}

		batch := append([]string(nil), files[start:end]...)
		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()

			args := append(append([]string(nil), baseArgs...), batch...)
			cmd := exec.Command(rgPath, args...)
			output, err := cmd.Output()

			found := false
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					if exitErr.ExitCode() == 1 {
						// No matches for this batch; still process the output below.
					} else {
						errChan <- fmt.Errorf("error executing ripgrep: %w", err)
						return
					}
				} else {
					errChan <- fmt.Errorf("error executing ripgrep: %w", err)
					return
				}
			} else {
				found = true
			}

			if !found && bytes.Contains(output, []byte(`"type":"match"`)) {
				found = true
			}

			if consume != nil {
				if err := consume(output); err != nil {
					errChan <- err
					return
				}
			}

			if found {
				matchesMu.Lock()
				matchesFound = true
				matchesMu.Unlock()
			}
		}(batch)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return matchesFound, err
		}
	}

	return matchesFound, nil
}
