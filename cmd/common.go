package cmd

import (
	"fmt"
	"os/exec"
	"strings"
)

func getPDFPageCount(pdfPath string) (int, error) {
	// Run the pdfinfo command to get the total number of pages
	cmd := exec.Command("pdfinfo", pdfPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to run pdfinfo: %v", err)
	}

	// Parse the output to find the "Pages" line
	var totalPages int
	lines := string(output)
	for _, line := range strings.Split(lines, "\n") {
		if strings.HasPrefix(line, "Pages:") {
			_, err := fmt.Sscanf(line, "Pages: %d", &totalPages)
			if err != nil {
				return 0, fmt.Errorf("failed to parse page count: %v", err)
			}
			break
		}
	}

	return totalPages, nil
}
func extractPDFPageWithPdftotext(pdfPath, outputPath string, page int) error {
	// Run the pdftotext command for a specific page
	cmd := exec.Command("pdftotext", "-f", fmt.Sprintf("%d", page), "-l", fmt.Sprintf("%d", page), pdfPath, outputPath)
	cmd.Stdout = nil
	cmd.Stderr = nil

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to extract text using pdftotext: %v", err)
	}

	return nil
}
func normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove leading and trailing spaces
	text = strings.TrimSpace(text)

	// Remove special characters (e.g., *, •, etc.)
	text = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == ' ' {
			return r
		}
		return -1
	}, text)

	// Replace multiple spaces with a single space
	text = strings.Join(strings.Fields(text), " ")

	return text
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
