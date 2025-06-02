package services

import (
	"fmt"
	"os"
	"os/exec"
	"pdf-extractor/internal/utils"
	"strings"

	"github.com/sirupsen/logrus"
)

func DeletePages(file string, fromPage int, toPage int, atPage int, startsWith string, backupPath string, backupFlag bool) error {
	// Implement the logic to delete pages from the PDF file
	// This function should handle the deletion of pages based on the provided parameters
	// and create a backup of the original file if backupPath is specified.

	err := validate(fromPage, toPage, atPage, startsWith)
	if err != nil {
		return err
	}
	err = utils.CreateDirectoryIfNotExists(backupPath)
	if err != nil {
		return err
	}
	if backupFlag {
		err = utils.CreateBackup(file, backupPath)
		if err != nil {
			return err
		}
	}
	// Proceed with the delete logic
	if atPage > 0 {
		// Call deletePageAt function
		deletePageAt(file, atPage)
	} else if startsWith != "" {
		// Call deletePagesByContent function
		deletePagesByContent(file, startsWith, toPage)
	} else if fromPage > 0 {
		// Call deletePagesRange function
		deletePagesRange(file, fromPage, toPage)
	}
	return nil
}

func validate(fromPage int, toPage int, atPage int, startsWith string) error {
	if atPage > 0 {
		if fromPage > 0 || toPage > 0 || startsWith != "" {
			return fmt.Errorf("error: --at flag cannot be used with --from, --to, or --starts-with")
		}
	} else if startsWith != "" {
		if atPage > 0 || fromPage > 0 {
			return fmt.Errorf("error: --starts-with flag cannot be used with --at or --from")
		}
		// If --to is not provided, set it to a very large number (end of the PDF)
		if toPage == 0 {
			toPage = int(^uint(0) >> 1) // Set to a very large number (max int value)
		}
	} else if fromPage > 0 {
		if atPage > 0 || startsWith != "" {
			return fmt.Errorf("error: --from flag cannot be used with --at or --starts-with")
		}
		if toPage == 0 {
			toPage = int(^uint(0) >> 1) // Set to a very large number (max int value)
		}
	} else {
		return fmt.Errorf("error: You must specify one of the following flags: --at, --from, or --starts-with")
	}
	return nil
}

func deletePageAt(pdfPath string, page int) error {
	// Get the total number of pages in the PDF
	totalPages, err := utils.GetPDFPageCount(pdfPath)
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
	totalPages, err := utils.GetPDFPageCount(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to get page count: %v", err)
	}

	// Validate that 'to' is within the valid range
	if to > totalPages || to == int(^uint(0)>>1) {
		to = totalPages // Set 'to' to the last page if it's very large
	}

	// Debug log: Total pages and 'to' value
	logrus.Debugf("[DEBUG] Total pages: %d, 'to' value: %d", totalPages, to)

	// Iterate through each page to find the page that starts with the specified string
	startPage := -1
	for page := 1; page <= totalPages; page++ {
		// Extract the content of the current page
		tempTxt := fmt.Sprintf("page_%d.txt", page)
		err := utils.ExtractPDFPageWithPdftotext(pdfPath, tempTxt, page)
		if err != nil {
			logrus.Errorf("Failed to extract page %d: %v", page, err)
			continue
		}

		// Read the content of the temporary file
		content, err := os.ReadFile(tempTxt)
		if err != nil {
			logrus.Errorf("Failed to read content of page %d: %v\n", page, err)
			os.Remove(tempTxt) // Clean up the temporary file
			continue
		}
		os.Remove(tempTxt) // Clean up the temporary file

		// Normalize the content and compare with 'startsWith'
		normalizedContent := utils.NormalizeText(string(content))
		normalizedStartsWith := utils.NormalizeText(startsWith)

		// Debug log: Print the starting words of the page
		logrus.Debugf("[DEBUG] Page %d starts with: '%s'", page, normalizedContent[:min(len(normalizedContent), 50)])

		// Check if the content starts with the specified string
		if strings.HasPrefix(normalizedContent, normalizedStartsWith) {
			startPage = page
			logrus.Debugf("[DEBUG] Found 'starts-with' match on page %d: '%s'", page, normalizedContent[:min(len(normalizedContent), 50)])
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
	logrus.Infof("[INFO] Successfully deleted pages starting from %d to %d in '%s'.", startPage, to, pdfPath)
	return nil
}

func deletePagesRange(pdfPath string, from, to int) error {
	// Get the total number of pages in the PDF
	totalPages, err := utils.GetPDFPageCount(pdfPath)
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
	logrus.Infof("[INFO] Successfully deleted pages from %d to %d in '%s'.", from, to, pdfPath)
	return nil
}
