package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

var undoFile string
var undoBackupPath string
var undoCmd = &cobra.Command{
	Use:   "undo-pdf",
	Short: "Undo the last modification to a PDF file",
	Long: `The undo-pdf command restores the latest backup of the specified PDF file.
It searches for the latest backup in the directory specified by the --backup-path flag (default: ./backup/<file-name>/).
The command replaces the current file with the latest backup and deletes the backup file after restoration.

You can specify a custom backup directory using the --backup-path flag:
Example: pdf-extractor undo-pdf --file="example.pdf" --backup-path="./custom_backup"`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate the --file flag
		if undoFile == "" {
			fmt.Println("Error: Please specify a file using the --file flag.")
			os.Exit(1)
		}

		// Perform the undo operation
		err := undoPDF(undoFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	undoCmd.Flags().StringVarP(&undoFile, "file", "f", "", "Path to the PDF file to undo")
	undoCmd.MarkFlagRequired("file")

	// Add --backup-path flag with default value "./backup"
	undoCmd.Flags().StringVar(&undoBackupPath, "backup-path", "./backup", "Path to the backup directory")

	rootCmd.AddCommand(undoCmd)
}

// undoPDF performs the undo operation
func undoPDF(filePath string) error {
	// Use the undoBackupPath variable for the backup directory
	fileName := filepath.Base(filePath)
	fileBackupDir := filepath.Join(undoBackupPath, fileName)

	// Check if the backup directory exists
	if _, err := os.Stat(fileBackupDir); os.IsNotExist(err) {
		return fmt.Errorf("no backup directory found for file: %s", fileName)
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
	err = copyFile(latestBackup, filePath)
	if err != nil {
		return fmt.Errorf("failed to restore the latest backup: %v", err)
	}

	// Verify if the backup file exists before deleting
	if _, err := os.Stat(latestBackup); os.IsNotExist(err) {
		return fmt.Errorf("backup file does not exist: %s", latestBackup)
	}

	// Delete the latest backup file
	err = os.Remove(latestBackup)
	if err != nil {
		return fmt.Errorf("failed to delete the latest backup file: %v", err)
	}

	fmt.Printf("Successfully restored the latest backup for '%s' and deleted the backup file.\n", filePath)
	return nil
}
