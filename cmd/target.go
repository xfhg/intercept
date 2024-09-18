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

			hash, err := calculateSHA256(path)
			if err != nil {
				return err
			}
			fileInfos = append(fileInfos, FileInfo{
				Path: path,
				Hash: hash,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking through directory: %w", err)
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

// func isIgnored(ignorePaths []string, path string) bool {
// 	absPath, err := filepath.Abs(path)
// 	if err != nil {
// 		// If the path cannot be converted to an absolute path, assume it's not ignored
// 		return false
// 	}

// 	// Convert all ignorePaths to absolute paths
// 	for _, ignore := range ignorePaths {
// 		absIgnore, err := filepath.Abs(ignore)
// 		if err != nil {
// 			continue // If an ignore path can't be converted, skip it
// 		}

// 		// Check if the absPath is a subpath of any absIgnore
// 		if strings.HasPrefix(absPath, absIgnore) {
// 			return true
// 		}
// 	}

// 	return false
// }

// func isIgnored(ignorePaths []string, path string) bool {
// 	absPath, err := filepath.Abs(path)
// 	if err != nil {
// 		// If the path cannot be converted to an absolute path, assume it's not ignored
// 		return false
// 	}

// 	// Convert all ignorePaths to absolute paths and handle extensions
// 	for _, ignore := range ignorePaths {
// 		// Check if the ignore pattern is an extension (e.g., *.md)
// 		if strings.HasPrefix(ignore, "*.") {
// 			// Extract the extension and compare it case-insensitively
// 			ext := strings.ToLower(ignore[1:]) // ".md" for "*.md"
// 			if strings.EqualFold(filepath.Ext(absPath), ext) {
// 				return true
// 			}
// 		} else {
// 			// Convert ignore to absolute path and compare paths
// 			absIgnore, err := filepath.Abs(ignore)
// 			if err != nil {
// 				continue // If an ignore path can't be converted, skip it
// 			}

// 			// Check if the absPath is a subpath of any absIgnore
// 			if strings.HasPrefix(absPath, absIgnore) {
// 				return true
// 			}
// 		}
// 	}

// 	return false
// }

func isIgnored(ignorePaths []string, path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If the path cannot be converted to an absolute path, assume it's not ignored
		return false
	}

	// Loop over each ignore pattern in ignorePaths
	for _, ignore := range ignorePaths {
		// Handle file extension patterns like *.md
		if strings.HasPrefix(ignore, "*.") {
			ext := strings.ToLower(ignore[1:]) // ".md" for "*.md"
			if strings.EqualFold(filepath.Ext(absPath), ext) {
				return true
			}
		} else if strings.HasSuffix(ignore, "/") {
			// Handle folder ignore patterns (like folder/) at any level
			absIgnore, err := filepath.Abs(ignore)
			if err != nil {
				continue // Skip if the ignore path cannot be resolved
			}

			// Check if we're ignoring this folder globally
			if strings.Contains(absPath, absIgnore) {
				return true
			}
		} else if strings.HasPrefix(ignore, "/") && strings.HasSuffix(ignore, "/") {
			// Handle root-level folder ignore (like /folder/)
			rootIgnore := filepath.Clean(ignore) // Clean the path
			if strings.HasPrefix(absPath, rootIgnore) {
				return true
			}
		} else {
			// Convert ignore to absolute path and compare paths
			absIgnore, err := filepath.Abs(ignore)
			if err != nil {
				continue // Skip if the ignore path cannot be resolved
			}

			// Check if the absPath is a subpath of any absIgnore
			if strings.HasPrefix(absPath, absIgnore) {
				return true
			}
		}
	}

	return false
}
