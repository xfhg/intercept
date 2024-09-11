//go:build !windows
// +build !windows

package cmd

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	maxFolderSize    = 100 * 1024 * 1024   // 100 MB
	compressionAge   = 24 * time.Hour      // Compress files older than 24 hours
	deletionAge      = 30 * 24 * time.Hour // Delete files older than 30 days
	compressedSuffix = ".gz"
)

func manageStatusReports() error {
	// Ensure the status directory exists
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create status directory: %w", err)
	}

	// Get all files in the status directory
	files, err := os.ReadDir(reportDir)
	if err != nil {
		return fmt.Errorf("failed to read status directory: %w", err)
	}

	var totalSize int64
	var fileInfos []os.FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			log.Warn().Err(err).Str("file", file.Name()).Msg("Failed to get file info")
			continue
		}
		totalSize += info.Size()
		fileInfos = append(fileInfos, info)
	}

	// Sort files by modification time (oldest first)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].ModTime().Before(fileInfos[j].ModTime())
	})

	now := time.Now()

	// Compress old files
	for _, info := range fileInfos {
		if now.Sub(info.ModTime()) > compressionAge && !strings.HasSuffix(info.Name(), compressedSuffix) {
			if err := compressFile(filepath.Join(reportDir, info.Name())); err != nil {
				log.Warn().Err(err).Str("file", info.Name()).Msg("Failed to compress file")
			} else {
				// Update total size after compression
				compressedInfo, err := os.Stat(filepath.Join(reportDir, info.Name()+compressedSuffix))
				if err == nil {
					totalSize = totalSize - info.Size() + compressedInfo.Size()
				}
			}
		}
	}

	// Delete old files if total size exceeds the limit
	for _, info := range fileInfos {
		if totalSize <= maxFolderSize {
			break
		}
		if now.Sub(info.ModTime()) > deletionAge {
			filePath := filepath.Join(reportDir, info.Name())
			if err := os.Remove(filePath); err != nil {
				log.Warn().Err(err).Str("file", info.Name()).Msg("Failed to delete old file")
			} else {
				totalSize -= info.Size()
				log.Info().Str("file", info.Name()).Msg("Deleted old report file")
			}
		}
	}

	return nil
}

func compressFile(filePath string) error {
	inputFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for compression: %w", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(filePath + compressedSuffix)
	if err != nil {
		return fmt.Errorf("failed to create compressed file: %w", err)
	}
	defer outputFile.Close()

	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, inputFile)
	if err != nil {
		return fmt.Errorf("failed to compress file: %w", err)
	}

	if err = os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove original file after compression: %w", err)
	}

	log.Info().Str("file", filepath.Base(filePath)).Msg("Compressed and removed original file")
	return nil
}
