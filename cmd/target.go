package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charlievieth/fastwalk"
)

type FileInfo struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

// CalculateFileHashes recursively calculates SHA256 hashes for all files in the given directory
func CalculateFileHashes(targetDir string) ([]FileInfo, error) {

	var fileInfos []FileInfo

	ignorePaths := policyData.Config.Flags.Ignore

	err := fastwalk.Walk(nil, targetDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && !isIgnored(ignorePaths, path) {

			if debugOutput {
				hash, err := calculateSHA256(path)
				if err != nil {
					return err
				}
				fileInfos = append(fileInfos, FileInfo{
					Path: path,
					Hash: hash,
				})
			} else {

				fileInfos = append(fileInfos, FileInfo{
					Path: path,
					Hash: "",
				})

			}

		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through directory: %w", err)
	}

	if len(fileInfos) == 0 {
		return nil, fmt.Errorf("no target files found at : %s", targetDir)
	}

	return fileInfos, nil
}

// WriteHashesToJSON writes the file hashes to a JSON file
func WriteHashesToJSON(fileInfos []FileInfo, outputPath string) error {
	if !debugOutput {
		return nil
	}

	outputPath = filepath.Join(outputDir, "_debug", outputPath)

	jsonData, err := json.MarshalIndent(fileInfos, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("error marshaling JSON")
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	err = os.WriteFile(outputPath, jsonData, 0644)
	if err != nil {
		log.Error().Err(err).Msg("error writing JSON file")
		return fmt.Errorf("error writing JSON file: %w", err)
	}

	return nil
}

// FilterFilesByPattern filters files based on a regex pattern
func FilterFilesByPattern(fileInfos []FileInfo, pattern string) ([]FileInfo, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var filteredFiles []FileInfo
	for _, fi := range fileInfos {
		if regex.MatchString(fi.Path) {
			filteredFiles = append(filteredFiles, fi)
		}
	}

	return filteredFiles, nil
}

func isIgnored(ignorePaths []string, path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If the path cannot be converted to an absolute path, assume it's not ignored
		return false
	}

	// Normalize the path separators
	absPath = filepath.ToSlash(absPath)

	for _, ignore := range ignorePaths {
		ignore = filepath.ToSlash(ignore)

		// Handle file extension patterns like *.md
		if strings.HasPrefix(ignore, "*.") {
			ext := strings.ToLower(ignore[1:]) // ".md" for "*.md"
			if strings.EqualFold(filepath.Ext(absPath), ext) {
				return true
			}
		} else if strings.HasSuffix(ignore, "/") {
			// Handle directory ignore patterns (like dist/)
			ignoreDir := ignore[:len(ignore)-1] // Remove trailing slash
			pathParts := strings.Split(absPath, "/")
			for _, part := range pathParts {
				if part == ignoreDir {
					return true
				}
			}
		} else {
			// For other patterns, check if the path contains the ignore pattern
			if strings.Contains(absPath, ignore) {
				return true
			}
		}
	}

	return false
}

func processIgnorePatterns(ignorePatterns []string) []string {
	processedPatterns := make([]string, 0, len(ignorePatterns)*2)
	for _, pattern := range ignorePatterns {
		processedPatterns = append(processedPatterns, "-g", fmt.Sprintf("!%s", pattern))
	}
	return processedPatterns
}

func detectOverlap(paths []string, targetPath string) (bool, string) {
	tp := normalizeCacheKey(targetPath)

	for _, p := range paths {
		pn := normalizeCacheKey(p)

		// exact match
		if pn == tp {
			return true, p
		}

		// directory existing -> target inside existing
		pnDir := normalizeDirectoryKey(pn)
		if strings.HasPrefix(tp, pnDir) {
			return true, p
		}

		// target is directory -> existing inside target
		tpDir := normalizeDirectoryKey(tp)
		if strings.HasPrefix(pn, tpDir) {
			return true, targetPath
		}
	}

	return false, ""
}
