package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var deleteFile string
var fromPage, toPage int
var atPage int
var startsWith string

var deleteCmd = &cobra.Command{
	Use:   "delete-pages",
	Short: "Delete pages from a PDF file",
	Long: `The delete-pages command allows you to delete specific pages or ranges of pages from a PDF file.
You can use one of the following options:

1. Delete a specific page:
   Use the --at flag to specify the page number to delete.
   Example: pdf-extractor delete-pages --file="example.pdf" --at=5

2. Delete a range of pages:
   Use the --from and --to flags to specify the starting and ending page numbers.
   If --to is not provided, all pages from the starting page to the end of the document will be deleted.
   Example: pdf-extractor delete-pages --file="example.pdf" --from=3 --to=7
   Example: pdf-extractor delete-pages --file="example.pdf" --from=10

3. Delete pages based on content:
   Use the --starts-with flag to delete pages where the content starts with the specified string.
   Example: pdf-extractor delete-pages --file="example.pdf" --starts-with="Introduction"

Note:
- You cannot combine --at, --from/--to, and --starts-with flags in a single command.
- The --file flag is required to specify the PDF file to operate on.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validation and delete logic here
		if atPage > 0 {
			if fromPage > 0 || toPage > 0 || startsWith != "" {
				fmt.Println("Error: --at flag cannot be used with --from, --to, or --starts-with.")
				os.Exit(1)
			}
		} else if startsWith != "" {
			if atPage > 0 || fromPage > 0 {
				fmt.Println("Error: --starts-with flag cannot be used with --at or --from.")
				os.Exit(1)
			}
			// If --to is not provided, set it to a very large number (end of the PDF)
			if toPage == 0 {
				toPage = int(^uint(0) >> 1) // Set to a very large number (max int value)
			}
		} else if fromPage > 0 {
			if atPage > 0 || startsWith != "" {
				fmt.Println("Error: --from flag cannot be used with --at or --starts-with.")
				os.Exit(1)
			}
			if toPage == 0 {
				toPage = int(^uint(0) >> 1) // Set to a very large number (max int value)
			}
		} else {
			fmt.Println("Error: You must specify one of the following flags: --at, --from, or --starts-with.")
			os.Exit(1)
		}
		if deleteFile == "" {
			fmt.Println("Error: Please specify a file using the --file flag.")
			os.Exit(1)
		}
		// Create a backup before deleting pages
		err := createBackup(deleteFile)
		if err != nil {
			fmt.Printf("Error creating backup: %v\n", err)
			os.Exit(1)
		}

		// Proceed with the delete logic
		if atPage > 0 {
			// Call deletePageAt function
			deletePageAt(deleteFile, atPage)
		} else if startsWith != "" {
			// Call deletePagesByContent function
			deletePagesByContent(deleteFile, startsWith, toPage)
		} else if fromPage > 0 {
			// Call deletePagesRange function
			deletePagesRange(deleteFile, fromPage, toPage)
		}
	},
}

func init() {
	deleteCmd.Flags().StringVarP(&deleteFile, "file", "f", "", "Path to the PDF file")
	deleteCmd.Flags().IntVar(&fromPage, "from", 0, "Starting page number to delete")
	deleteCmd.Flags().IntVar(&toPage, "to", 0, "Ending page number to delete")
	deleteCmd.Flags().IntVar(&atPage, "at", 0, "Specific page number to delete")
	deleteCmd.Flags().StringVar(&startsWith, "starts-with", "", "Delete pages where content starts with the specified string")
	deleteCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(deleteCmd)
}

func createBackup(pdfPath string) error {
	// Get the backup directory path
	backupDir := "backup"
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		err := os.Mkdir(backupDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create backup directory: %v", err)
		}
	}

	// Create a subdirectory for the specific PDF file
	pdfName := filepath.Base(pdfPath)
	pdfBackupDir := filepath.Join(backupDir, pdfName)
	if _, err := os.Stat(pdfBackupDir); os.IsNotExist(err) {
		err := os.Mkdir(pdfBackupDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create PDF-specific backup directory: %v", err)
		}
	}

	// Create a backup file with a timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFile := filepath.Join(pdfBackupDir, fmt.Sprintf("%s_%s.pdf", pdfName, timestamp))
	err := copyFile(pdfPath, backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Enforce global LRU strategy
	err = enforceGlobalLRU(backupDir)
	if err != nil {
		return fmt.Errorf("failed to enforce global LRU strategy: %v", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	return err
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

func deletePagesRange(pdfPath string, from, to int) error {
	// Get the total number of pages in the PDF
	totalPages, err := getPDFPageCount(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to get page count: %v", err)
	}

	// If 'to' is very large, set it to the total number of pages
	if to > totalPages || to == int(^uint(0)>>1) {
		to = totalPages
	}

	// Validate the range
	if from < 1 || from > totalPages || to < from {
		return fmt.Errorf("invalid page range: from=%d, to=%d, totalPages=%d", from, to, totalPages)
	}

	// Construct the page ranges to keep
	var pagesToKeep []string
	if from > 1 {
		pagesToKeep = append(pagesToKeep, fmt.Sprintf("1-%d", from-1))
	}
	if to < totalPages {
		pagesToKeep = append(pagesToKeep, fmt.Sprintf("%d-%d", to+1, totalPages))
	}

	// Use pdftk to create a new PDF with the remaining pages
	tempFile := "temp.pdf"
	cmdArgs := append([]string{pdfPath, "cat"}, pagesToKeep...)
	cmdArgs = append(cmdArgs, "output", tempFile)
	cmd := exec.Command("pdftk", cmdArgs...)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to delete pages using pdftk: %v", err)
	}

	// Overwrite the original PDF with the new one
	err = os.Rename(tempFile, pdfPath)
	if err != nil {
		return fmt.Errorf("failed to overwrite original PDF: %v", err)
	}

	fmt.Printf("Successfully deleted pages from %d to %d in '%s'.\n", from, to, pdfPath)
	return nil
}

func deletePageAt(pdfPath string, page int) error {
	// Get the total number of pages in the PDF
	totalPages, err := getPDFPageCount(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to get page count: %v", err)
	}

	// Validate if the page lies within the valid range
	if page < 1 || page > totalPages {
		return fmt.Errorf("invalid page number: %d (total pages: %d)", page, totalPages)
	}

	// Construct the page ranges to keep
	var pagesToKeep []string
	if page > 1 {
		pagesToKeep = append(pagesToKeep, fmt.Sprintf("1-%d", page-1))
	}
	if page < totalPages {
		pagesToKeep = append(pagesToKeep, fmt.Sprintf("%d-%d", page+1, totalPages))
	}

	// Use pdftk to create a new PDF with the remaining pages
	tempFile := "temp.pdf"
	cmdArgs := append([]string{pdfPath, "cat"}, pagesToKeep...)
	cmdArgs = append(cmdArgs, "output", tempFile)
	cmd := exec.Command("pdftk", cmdArgs...)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to delete page using pdftk: %v", err)
	}

	// Overwrite the original PDF with the new one
	err = os.Rename(tempFile, pdfPath)
	if err != nil {
		return fmt.Errorf("failed to overwrite original PDF: %v", err)
	}

	fmt.Printf("Successfully deleted page %d in '%s'.\n", page, pdfPath)
	return nil
}

func deletePagesByContent(pdfPath, startsWith string, to int) error {
	// Get the total number of pages in the PDF
	totalPages, err := getPDFPageCount(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to get page count: %v", err)
	}

	// Validate that 'to' is within the valid range
	if to > totalPages || to == int(^uint(0)>>1) {
		to = totalPages // Set 'to' to the last page if it's very large
	}

	// Debug log: Total pages and 'to' value
	fmt.Printf("[DEBUG] Total pages: %d, 'to' value: %d\n", totalPages, to)

	// Iterate through each page to find the page that starts with the specified string
	startPage := -1
	for page := 1; page <= totalPages; page++ {
		// Extract the content of the current page
		tempTxt := fmt.Sprintf("page_%d.txt", page)
		err := extractPDFPageWithPdftotext(pdfPath, tempTxt, page)
		if err != nil {
			fmt.Printf("[DEBUG] Failed to extract page %d: %v\n", page, err)
			continue
		}

		// Read the content of the temporary file
		content, err := os.ReadFile(tempTxt)
		if err != nil {
			fmt.Printf("[DEBUG] Failed to read content of page %d: %v\n", page, err)
			os.Remove(tempTxt) // Clean up the temporary file
			continue
		}
		os.Remove(tempTxt) // Clean up the temporary file

		// Normalize the content and compare with 'startsWith'
		normalizedContent := normalizeText(string(content))
		normalizedStartsWith := normalizeText(startsWith)

		// Debug log: Print the starting words of the page
		fmt.Printf("[DEBUG] Page %d starts with: '%s'\n", page, normalizedContent[:min(len(normalizedContent), 50)])

		// Check if the content starts with the specified string
		if strings.HasPrefix(normalizedContent, normalizedStartsWith) {
			startPage = page
			fmt.Printf("[DEBUG] Found 'starts-with' match on page %d\n", page)
			break
		}
	}

	// Validate that the 'starts-with' string was found
	if startPage == -1 {
		return fmt.Errorf("no page starts with the specified string: '%s'", startsWith)
	}

	// Validate that 'to' is greater than or equal to the startPage
	if to < startPage {
		return fmt.Errorf("'to' (%d) must be greater than or equal to the page where 'starts-with' occurs (%d)", to, startPage)
	}

	// Construct the page ranges to keep
	var pagesToKeep []string
	if startPage > 1 {
		pagesToKeep = append(pagesToKeep, fmt.Sprintf("1-%d", startPage-1))
	}
	if to < totalPages {
		pagesToKeep = append(pagesToKeep, fmt.Sprintf("%d-%d", to+1, totalPages))
	}

	// Use pdftk to create a new PDF with the remaining pages
	tempFile := "temp.pdf"
	cmdArgs := append([]string{pdfPath, "cat"}, pagesToKeep...)
	cmdArgs = append(cmdArgs, "output", tempFile)
	cmd := exec.Command("pdftk", cmdArgs...)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to delete pages using pdftk: %v", err)
	}

	// Overwrite the original PDF with the new one
	err = os.Rename(tempFile, pdfPath)
	if err != nil {
		return fmt.Errorf("failed to overwrite original PDF: %v", err)
	}

	fmt.Printf("Successfully deleted pages from %d to %d in '%s'.\n", startPage, to, pdfPath)
	return nil
}
