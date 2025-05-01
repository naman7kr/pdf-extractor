package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var extractFile string

var extractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract pages related to articles from the PDF",
	Run:   processExtract,
}

func init() {
	extractCmd.Flags().StringVarP(&extractFile, "file", "f", "", "Path to the PDF file")
	extractCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(extractCmd)
}

func processExtract(cmd *cobra.Command, args []string) {
	if extractFile == "" {
		fmt.Println("Error: Please specify a file using the --file flag.")
		os.Exit(1)
	}

	// Check if articles.txt exists
	articlesFile := "articles.txt"
	if _, err := os.Stat(articlesFile); os.IsNotExist(err) {
		fmt.Printf("Error: %s not found. Please run the 'chapters' command first to generate it.\n", articlesFile)
		os.Exit(1)
	}

	// Read articles from articles.txt
	articles, err := readArticlesFromFile(articlesFile)
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

	// Create the "extracted" folder if it doesn't exist
	extractedFolder := "extracted"
	err = os.MkdirAll(extractedFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create 'extracted' folder: %v", err)
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

		// Log the starting text of the page for debugging
		fmt.Printf("Page %d starting text: '%s'\n", page, normalizedContent[:min(len(normalizedContent), 100)]) // Log first 100 characters

		// Check if the first k characters match any article title
		for _, article := range articles {
			if matchArticleTitleByLength(normalizedContent, article) {
				fmt.Printf("Matched article '%s' on page %d\n", article, page) // Debugging log
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
			endPage = totalPages
		}

		// Validate page range
		if startPage > endPage || startPage < 1 || endPage > totalPages {
			fmt.Printf("Warning: Invalid page range for article '%s' (start: %d, end: %d).\n", article, startPage, endPage)
			continue
		}

		// Generate the output file path
		outputFile := filepath.Join(extractedFolder, fmt.Sprintf("%s.pdf", sanitizeFileName(article)))

		// Extract the pages for the current article
		err := extractPDFPages(pdfPath, outputFile, startPage, endPage)
		if err != nil {
			return fmt.Errorf("failed to extract pages for article '%s': %v", article, err)
		}

		fmt.Printf("Extracted pages %d to %d for article '%s' into '%s'\n", startPage, endPage, article, outputFile)
	}

	return nil
}

func matchArticleTitleByLength(content, article string) bool {
	// Normalize both the content and the article title
	normalizedContent := normalizeText(content)
	normalizedArticle := normalizeText(article)

	// Ensure the content is long enough to compare
	if len(normalizedContent) < len(normalizedArticle) {
		return false
	}

	// Compare the normalized article title with a substring of the normalized content
	return strings.Contains(normalizedContent, normalizedArticle)
}

func sanitizeFileName(name string) string {
	// Replace invalid characters in the filename
	return strings.ReplaceAll(name, " ", "_")
}

func extractPDFPages(pdfPath, outputPath string, startPage, endPage int) error {
	// Run the pdftk command to extract pages
	cmd := exec.Command("pdftk", pdfPath, "cat", fmt.Sprintf("%d-%d", startPage, endPage), "output", outputPath)
	cmd.Stdout = nil
	cmd.Stderr = nil

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to extract pages using pdftk: %v", err)
	}

	return nil
}
