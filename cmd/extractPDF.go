package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var extractFile string
var configPath string
var endsWith string
var chapterOutputPath string
var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract pages related to articles from the PDF",
	Run:   processExtract,
}

func init() {
	extractCmd.Flags().StringVarP(&extractFile, "file", "f", "", "Path to the PDF file")
	extractCmd.MarkFlagRequired("file")

	// Add --output-path flag with default value "extracted/"
	extractCmd.Flags().StringVarP(&chapterOutputPath, "output-path", "o", "./extracted", "Path where the PDF files are generated")

	// Add --config-path flag with default value "./articles.txt"
	extractCmd.Flags().StringVarP(&configPath, "config-path", "c", "./articles.txt", "Path to the .txt file where articles are listed")

	// Add --ends-with flag
	extractCmd.Flags().StringVar(&endsWith, "ends-with", "", "Text to find the page where the last article ends")

	rootCmd.AddCommand(extractCmd)
}

func processExtract(cmd *cobra.Command, args []string) {
	if extractFile == "" {
		fmt.Println("Error: Please specify a file using the --file flag.")
		os.Exit(1)
	}
	// Ensure chapterOutputPath has the default value if not provided
	fmt.Println("Output path: ", chapterOutputPath)
	// Ensure the output directory exists
	err := os.MkdirAll(chapterOutputPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Check if the articles file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Error: %s not found. Please provide a valid articles file.\n", configPath)
		os.Exit(1)
	}

	// Read articles from the specified config file
	articles, err := readArticlesFromFile(configPath)
	if err != nil {
		fmt.Printf("Error reading articles: %v\n", err)
		os.Exit(1)
	}

	// Extract pages for each article
	err = extractPagesForArticles(extractFile, articles)
	if err != nil {
		fmt.Printf("Error extracting pages: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Pages successfully extracted for all articles.")
}

func readArticlesFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var articles []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		articles = append(articles, strings.TrimSpace(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return articles, nil
}

func extractPagesForArticles(pdfPath string, articles []string) error {
	// Get the total number of pages in the PDF
	totalPages, err := getPDFPageCount(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to get page count: %v", err)
	}

	// Ensure the output folder exists
	err = os.MkdirAll(chapterOutputPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output folder: %v", err)
	}

	// Map to store the starting page of each article
	articlePages := make(map[string]int)

	// Iterate through each page to find the starting page of each article
	for page := 1; page <= totalPages; page++ {
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("page_%d.txt", page))

		// Extract the current page using pdftotext
		err := extractPDFPageWithPdftotext(pdfPath, tempFile, page)
		if err != nil {
			return fmt.Errorf("failed to extract page %d: %v", page, err)
		}

		// Read the extracted content from the temporary file
		content, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("failed to read temporary file for page %d: %v", page, err)
		}

		// Normalize the extracted content by removing line breaks
		normalizedContent := normalizeText(string(content))

		// Check if the first k characters match any article title
		for _, article := range articles {
			if matchArticleTitleByLength(normalizedContent, article) {
				articlePages[article] = page
				break
			}
		}

		// Remove the temporary file
		os.Remove(tempFile)
	}

	// Extract pages for each article
	for i, article := range articles {
		startPage, ok := articlePages[article]
		if !ok {
			fmt.Printf("Warning: Article '%s' not found in the PDF.\n", article)
			continue
		}

		// Determine the end page
		var endPage int
		if i+1 < len(articles) {
			nextArticle := articles[i+1]
			if nextStartPage, ok := articlePages[nextArticle]; ok {
				endPage = nextStartPage - 1
			} else {
				endPage = totalPages
			}
		} else {
			// Handle the last article
			endPage = totalPages
			if endsWith != "" {
				pageFound, err := findPageEndingWith(pdfPath, endsWith, startPage, totalPages)
				if err != nil {
					fmt.Printf("Error finding page for --ends-with: %v\n", err)
				} else if pageFound > 0 {
					endPage = pageFound - 1
				}
			}
		}

		// Validate page range
		if startPage > endPage || startPage < 1 || endPage > totalPages {
			fmt.Printf("Warning: Invalid page range for article '%s' (start: %d, end: %d).\n", article, startPage, endPage)
			continue
		}

		// Generate the output file path using chapterOutputPath
		outputFile := filepath.Join(chapterOutputPath, fmt.Sprintf("%s.pdf", sanitizeFileName(article)))

		// Extract the pages for the current article
		err := extractPDFPages(pdfPath, outputFile, startPage, endPage)
		if err != nil {
			return fmt.Errorf("failed to extract pages for article '%s': %v", article, err)
		}

		fmt.Printf("Extracted pages %d to %d for article '%s' into '%s'\n", startPage, endPage, article, outputFile)
	}

	return nil
}

func findPageEndingWith(pdfPath, endsWith string, startPage, totalPages int) (int, error) {
	for page := startPage; page <= totalPages; page++ {
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("page_%d.txt", page))

		// Extract the current page using pdftotext
		err := extractPDFPageWithPdftotext(pdfPath, tempFile, page)
		if err != nil {
			return 0, fmt.Errorf("failed to extract page %d: %v", page, err)
		}

		// Read the extracted content from the temporary file
		content, err := os.ReadFile(tempFile)
		if err != nil {
			return 0, fmt.Errorf("failed to read temporary file for page %d: %v", page, err)
		}

		// Normalize the extracted content
		normalizedContent := normalizeText(string(content))

		// Check if the content starts with the specified text
		if strings.HasPrefix(normalizedContent, normalizeText(endsWith)) {
			return page, nil
		}

		// Remove the temporary file
		os.Remove(tempFile)
	}

	return 0, nil
}
func matchArticleTitleByLength(content, article string) bool {
	// Normalize both the content and the article title
	normalizedContent := normalizeText(content)
	normalizedArticle := normalizeText(article)
	// if normalizedContent starts with "women in entrepreneurship"
	if strings.HasPrefix(normalizedContent, "women in entrepreneurship") {
		fmt.Println("Normalized Content: ", normalizedContent)
		fmt.Println("Normalized Article: ", normalizedArticle)
		fmt.Println("Content Length: ", len(normalizedContent))
		fmt.Println("Article Length: ", len(normalizedArticle))
	}
	// Ensure the content is long enough to compare
	if len(normalizedContent) < len(normalizedArticle) {
		return false
	}

	// Compare the normalized article title with a substring of the normalized content
	return strings.Contains(normalizedContent, normalizedArticle)
}

func sanitizeFileName(name string) string {
	// Replace spaces with underscores first
	name = strings.ReplaceAll(name, " ", "_")
	// Remove all special characters apart from underscores
	name = regexp.MustCompile(`[^a-zA-Z0-9_]`).ReplaceAllString(name, "")
	return name
}

func extractPDFPages(pdfPath, chapterOutputPath string, startPage, endPage int) error {
	// Run the pdftk command to extract pages
	cmd := exec.Command("pdftk", pdfPath, "cat", fmt.Sprintf("%d-%d", startPage, endPage), "output", chapterOutputPath)
	cmd.Stdout = nil
	cmd.Stderr = nil

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to extract pages using pdftk: %v", err)
	}

	return nil
}
