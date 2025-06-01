package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func CreateBackup(pdfPath string, backupPath string) error {
	// Use the backupPath variable for the backup directory
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		err := os.Mkdir(backupPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create backup directory: %v", err)
		}
	}

	// Create a subdirectory for the specific PDF file
	pdfName := filepath.Base(pdfPath)
	pdfBackupDir := filepath.Join(backupPath, strings.TrimSuffix(pdfName, ".pdf"))
	if _, err := os.Stat(pdfBackupDir); os.IsNotExist(err) {
		err := os.Mkdir(pdfBackupDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create PDF-specific backup directory: %v", err)
		}
	}

	// Create a backup file with a timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(pdfBackupDir, fmt.Sprintf("%s_%s.pdf", pdfName, timestamp))
	err := CopyFile(pdfPath, backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Enforce global LRU strategy
	err = enforceGlobalLRU(backupPath)
	if err != nil {
		return fmt.Errorf("failed to enforce global LRU strategy: %v", err)
	}

	return nil
}

func enforceGlobalLRU(backupDir string) error {
	// Get the TOTAL_CAPACITY environment variable or default to 20
	totalCapacityStr := os.Getenv("BACKUP_CAPACITY")
	totalCapacity := 20
	if totalCapacityStr != "" {
		if val, err := strconv.Atoi(totalCapacityStr); err == nil {
			totalCapacity = val
		}
	}

	// Get all backup files across all subdirectories
	var allFiles []string
	err := filepath.Walk(backupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Only include files (not directories)
		if !info.IsDir() {
			allFiles = append(allFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %v", err)
	}

	// If the total number of files exceeds the capacity, delete the oldest files
	if len(allFiles) > totalCapacity {
		// Sort files by modification time (oldest first)
		sort.Slice(allFiles, func(i, j int) bool {
			fileInfoI, _ := os.Stat(allFiles[i])
			fileInfoJ, _ := os.Stat(allFiles[j])
			return fileInfoI.ModTime().Before(fileInfoJ.ModTime())
		})

		// Delete the oldest files to maintain the capacity
		for i := 0; i < len(allFiles)-totalCapacity; i++ {
			err := os.Remove(allFiles[i])
			if err != nil {
				return fmt.Errorf("failed to delete old backup file: %v", err)
			}
		}
	}

	return nil
}
