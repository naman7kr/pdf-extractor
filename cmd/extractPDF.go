package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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

	// Update --config-path flag to point to a directory
	extractCmd.Flags().StringVarP(&configPath, "config-path", "c", "./configs", "Path to the directory containing the articles.txt file")

	// Add --ends-with flag
	extractCmd.Flags().StringVar(&endsWith, "ends-with", "", "Text to find the page where the last article ends")

	rootCmd.AddCommand(extractCmd)
}
func readArticlesFromConfig(filePath string) ([]string, error) {
	// Read the YAML file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %v", err)
	}

	// Parse the YAML file
	var config ArticlesConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config.yaml: %v", err)
	}

	// Extract article titles
	var titles []string
	for _, article := range config.Articles {
		titles = append(titles, article.Title)
	}

	return titles, nil
}
func processExtract(cmd *cobra.Command, args []string) {
	if extractFile == "" {
		fmt.Println("Error: Please specify a file using the --file flag.")
		os.Exit(1)
	}

	// Ensure the output directory exists
	err := os.MkdirAll(chapterOutputPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Construct the full path to the config.yaml file
	configFilePath := filepath.Join(configPath, "config.yaml")

	// Check if the config.yaml file exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		fmt.Printf("Error: %s not found. Please provide a valid directory containing config.yaml.\n", configFilePath)
		os.Exit(1)
	}

	// Read articles from the config.yaml file
	articles, err := readArticlesFromConfig(configFilePath)
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
	const patternThreshold = 0.4 // 80% threshold
	const patternLimit = 100
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

	// Array to store normalized content of all pages
	pageContents := make([]string, totalPages)

	// Iterate through each page to extract and store normalized content
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
		pageContents[page-1] = normalizedContent // Store normalized content in the array

		// Remove the temporary file
		os.Remove(tempFile)
	}
	longestPrefix, err := findLongestPrefix(pageContents, patternThreshold)
	if err != nil {
		fmt.Println("Error finding prefix:", err)
	} else {
		fmt.Println("Detected Longest Common Prefix:", longestPrefix)
	}

	pageContents = removePrefix(pageContents, longestPrefix)
	for i, content := range pageContents {
		// print 100 characters of the content for debugging
		if i+1 == 17 {
			fmt.Printf("Page %d content (first 100 chars): %s\n", i+1, normalizeText(content[:150]))
			fmt.Printf("article[2]: %s\n", normalizeText(articles[2]))
		}
	}
	// find starting pages for articles
	for _, article := range articles {
		for page, normalizedContent := range pageContents {

			if matchArticleTitleByLength(normalizedContent, article) {
				fmt.Printf("Found article '%s' on page %d\n", article, page+1)
				articlePages[article] = page + 1 // Pages are 1-indexed
				break
			}
		}
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

// removeDigits removes all digits from a given string.
func removeDigits(input string) string {
	re := regexp.MustCompile("[0-9]")
	return re.ReplaceAllString(input, "")
}

// findLongestPrefix finds the longest prefix from both the start and reverse order of the list.
func findLongestPrefix(stringsList []string, threshold float64) (string, error) {
	processedStrings := make([]string, len(stringsList))
	for i, str := range stringsList {
		processedStrings[i] = removeDigits(str)
	}

	totalStrings := len(processedStrings)
	if totalStrings == 0 {
		return "", fmt.Errorf("no strings provided")
	}

	// Helper function to find prefix in a given order
	findPrefix := func(strings []string) string {
		firstString := strings[0]
		longestPrefix := ""

		for i := 1; i <= len(firstString); i++ {
			prefix := firstString[:i]
			count := 0

			for _, str := range strings {
				if stringStartsWith(str, prefix) {
					count++
				}
			}

			occurrenceRate := float64(count) / float64(totalStrings)
			if occurrenceRate >= threshold && len(prefix) > len(longestPrefix) {
				longestPrefix = prefix
			}
		}
		return longestPrefix
	}

	// Find prefix from the start
	startPrefix := findPrefix(processedStrings)

	// Find prefix from the reverse order
	reversedStrings := make([]string, len(processedStrings))
	copy(reversedStrings, processedStrings)
	for i := 0; i < len(reversedStrings)/2; i++ {
		reversedStrings[i], reversedStrings[len(reversedStrings)-1-i] = reversedStrings[len(reversedStrings)-1-i], reversedStrings[i]
	}
	endPrefix := findPrefix(reversedStrings)

	// Return the longer of the two prefixes
	if len(startPrefix) > len(endPrefix) {
		return startPrefix, nil
	}
	return endPrefix, nil
}

// startsWith checks if a string starts with the given prefix.
func stringStartsWith(str, prefix string) bool {
	return len(str) >= len(prefix) && str[:len(prefix)] == prefix
}

// removePrefix removes the identified prefix from the original strings.
func removePrefix(stringsList []string, prefix string) []string {
	result := []string{}

	for _, str := range stringsList {
		// Remove digits from the string to check if the prefix exists
		processedStr := removeDigits(str)
		if stringStartsWith(processedStr, prefix) {
			// Iterate through the string and remove characters until the length of the prefix is reached
			prefixLength := len(prefix)
			removedLength := 0
			for i := 0; i < len(str); i++ {
				if removedLength >= prefixLength {
					str = str[i:]
					break
				}
				if !isDigit(str[i]) {
					removedLength++
				}
			}
			// If the length of the prefix is not reached, don't remove any characters
			if removedLength < prefixLength {
				str = stringsList[len(result)] // Restore original string
			}
		}
		result = append(result, str)
	}

	return result
}

// isDigit checks if a character is a digit.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
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

	// Ensure the content is long enough to compare
	if len(normalizedContent) < len(normalizedArticle) {
		return false
	}

	// Compare the normalized article title with a substring of the normalized content
	return strings.HasPrefix(normalizedContent, normalizedArticle)
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
