package services

import (
	"fmt"
	"os"
	"path/filepath"
	"pdf-extractor/internal/utils"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

func Undo(file string, backupPath string) error {
	fileName := filepath.Base(file)
	fileBackupDir := filepath.Join(backupPath, strings.TrimSuffix(fileName, ".pdf"))
	err := utils.CheckFileExists(fileBackupDir)
	if err != nil {
		return err
	}
	// Get all backup files in the directory
	files, err := os.ReadDir(fileBackupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %v", err)
	}
	// Check if there are any backup files
	if len(files) == 0 {
		return fmt.Errorf("no backup files found for file: %s", fileName)
	}

	// Sort the files by modification time (latest first)
	sort.Slice(files, func(i, j int) bool {
		fileInfoI, _ := files[i].Info()
		fileInfoJ, _ := files[j].Info()
		return fileInfoI.ModTime().After(fileInfoJ.ModTime())
	})

	// Get the latest backup file
	latestBackup := filepath.Join(fileBackupDir, files[0].Name())

	// Debug log: Print the latest backup file path
	fmt.Printf("[DEBUG] Latest backup file: %s\n", latestBackup)

	// Copy the latest backup to replace the current file
	err = utils.CopyFile(latestBackup, file)
	if err != nil {
		return fmt.Errorf("failed to restore the latest backup: %v", err)
	}

	// Verify if the backup file exists before deleting
	err = utils.CheckFileExists(latestBackup)

	if err != nil {
		return fmt.Errorf("backup file does not exist: %s", latestBackup)
	}

	// Delete the latest backup file
	err = os.Remove(latestBackup)
	if err != nil {
		return fmt.Errorf("failed to delete the latest backup file: %v", err)
	}
	logrus.Infof("Successfully restored the latest backup for '%s' and deleted the backup file.", file)
	return nil
}
