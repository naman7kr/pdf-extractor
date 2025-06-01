package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func CreateDirectoryIfNotExists(path string) error {
	// Check if the directory exists
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", path, err)
	}
	return nil
}
func CheckFileExists(filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", filePath) // Return a descriptive error
		}
		return err // Return the actual error if it's not a "file not found" error
	}
	return nil // File exists
}

func GetPDFPageCount(pdfPath string) (int, error) {
	// Run the pdfinfo command to get the total number of pages
	fmt.Println("PDF Path:", pdfPath)
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
func ExtractPDFPageWithPdftotext(pdfPath, outputPath string, page int) error {
	// Run the pdftotext command for a specific page in raw mode
	cmd := exec.Command("pdftotext", "-layout", "-f", fmt.Sprintf("%d", page), "-l", fmt.Sprintf("%d", page), pdfPath, outputPath)
	cmd.Stdout = nil
	cmd.Stderr = nil

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to extract text using pdftotext: %v", err)
	}

	return nil
}
func NormalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove leading and trailing spaces
	text = strings.TrimSpace(text)

	// Allow specific characters (a-z, 0-9, space, ',', ':', ';', '-')
	text = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, text)

	// Replace multiple spaces with a single space
	text = strings.Join(strings.Fields(text), " ")

	return text
}
func TrimTrailingNumber(title string) string {
	// Regular expression to match trailing numbers and spaces
	re := regexp.MustCompile(`\s*\d+$`)
	// Replace trailing numbers and spaces with an empty string
	return strings.TrimSpace(re.ReplaceAllString(title, ""))
}
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func SanitizeFileName(name string) string {
	// Replace spaces with underscores first
	name = strings.ReplaceAll(name, " ", "_")
	// Remove all special characters apart from underscores
	name = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(name, "")
	return name
}

func RemoveDigits(input string) string {
	re := regexp.MustCompile("[0-9]")
	return re.ReplaceAllString(input, "")
}
func IsDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func CopyFile(src, dst string) error {
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
func DeleteFile(filePath string) error {
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %v", filePath, err)
	}
	return nil
}
