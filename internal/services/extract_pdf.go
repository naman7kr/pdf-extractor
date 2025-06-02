package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"pdf-extractor/internal/models"
	"pdf-extractor/internal/utils"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func ExtractPDF(extractFile string, outputPath string, configPath string, endsWith string) error {

	err := utils.RecreateDirectory(outputPath)
	if err != nil {
		return err
	}

	configFilePath := filepath.Join(configPath, "config.yaml")
	err = utils.CheckFileExists(configFilePath)
	if err != nil {
		return err
	}
	// Read articles from the config.yaml file
	articles, err := readArticlesFromConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("error reading articles: %v", err)
	}

	// Extract pages for each article
	err = extractPagesForArticles(extractFile, articles, outputPath, endsWith)
	if err != nil {
		return fmt.Errorf("error extracting pages: %v", err)
	}

	logrus.Infof("Pages successfully extracted for all articles in %s", outputPath)

	return nil
}
func readArticlesFromConfig(filePath string) ([]string, error) {
	// Read the YAML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %v", err)
	}

	// Parse the YAML file
	var config models.ArticlesConfig
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

func extractPagesForArticles(pdfPath string, articles []string, outputPath string, endsWith string) error {
	outputFile := ""
	const patternThreshold = 0.6 // 80% threshold
	// Get the total number of pages in the PDF
	totalPages, err := utils.GetPDFPageCount(pdfPath)
	if err != nil {
		return fmt.Errorf("failed to get page count: %v", err)
	}

	// Map to store the starting page of each article
	articlePages := make(map[string]int)

	// Array to store normalized content of all pages
	pageContents := make([]string, totalPages)

	// Iterate through each page to extract and store normalized content
	for page := 1; page <= totalPages; page++ {
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("page_%d.txt", page))

		// Extract the current page using pdftotext
		err := utils.ExtractPDFPageWithPdftotext(pdfPath, tempFile, page)
		if err != nil {
			return fmt.Errorf("failed to extract page %d: %v", page, err)
		}

		// Read the extracted content from the temporary file
		content, err := os.ReadFile(tempFile)
		if err != nil {
			return fmt.Errorf("failed to read temporary file for page %d: %v", page, err)
		}

		// Normalize the extracted content by removing line breaks
		normalizedContent := utils.NormalizeText(string(content))
		pageContents[page-1] = normalizedContent // Store normalized content in the array

		// Remove the temporary file
		os.Remove(tempFile)
	}
	longestPrefix, err := findLongestPrefix(pageContents, patternThreshold)
	if err != nil {
		return fmt.Errorf("error finding prefix: %v", err)
	} else {
		logrus.Debugf("Longest prefix found: '%s'", longestPrefix)
	}

	pageContents = removePrefix(pageContents, longestPrefix)
	// find starting pages for articles
	for _, article := range articles {
		for page, normalizedContent := range pageContents {

			if matchArticleTitleByLength(normalizedContent, article) {
				logrus.Debugf("Found article '%s' on page %d", article, page+1)
				articlePages[article] = page + 1 // Pages are 1-indexed
				break
			}
		}
	}

	// Extract pages for each article
	for i, article := range articles {
		startPage, ok := articlePages[article]
		if !ok {
			logrus.Warnf("Article '%s' not found in the PDF.", article)
			continue
		}
		var endPage int

		// Determine the end page

		if i+1 < len(articles) {
			nextArticle := articles[i+1]
			if nextStartPage, ok := articlePages[nextArticle]; ok {
				endPage = nextStartPage - 1
			} else {
				endPage = totalPages
			}
		} else {
			// logrus.Warnln("This is the last article")
			// Handle the last article
			endPage = totalPages
			if endsWith != "" {
				pageFound, err := findPageEndingWith(pdfPath, endsWith, startPage, totalPages, pageContents)
				if err != nil {
					return fmt.Errorf("error finding page for --ends-with: %v", err)
				} else if pageFound > 0 {
					endPage = pageFound - 1
				}
			}
		}
		// logrus.Warnf("Endpage for article '%s' is %d", article, endPage)
		// Validate page range
		if startPage > endPage || startPage < 1 || endPage > totalPages {
			return fmt.Errorf("invalid page range for article '%s' (start: %d, end: %d)", article, startPage, endPage)
		}

		// Generate the output file path using chapterOutputPath
		outputFile = filepath.Join(outputPath, fmt.Sprintf("%s.pdf", utils.SanitizeFileName(article)))

		// Extract the pages for the current article
		err := extractPDFPages(pdfPath, outputFile, startPage, endPage)
		if err != nil {
			return fmt.Errorf("failed to extract pages for article '%s': %v", article, err)
		}
		logrus.Infof("Extracted pages %d to %d for article '%s' into '%s'", startPage, endPage, article, outputFile)
	}

	return nil
}

func findLongestPrefix(stringsList []string, threshold float64) (string, error) {
	processedStrings := make([]string, len(stringsList))
	for i, str := range stringsList {
		processedStrings[i] = utils.RemoveDigits(str)
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
		processedStr := utils.RemoveDigits(str)
		if stringStartsWith(processedStr, prefix) {
			// Iterate through the string and remove characters until the length of the prefix is reached
			prefixLength := len(prefix)
			removedLength := 0
			for i := 0; i < len(str); i++ {
				if removedLength >= prefixLength {
					str = str[i:]
					break
				}
				if !utils.IsDigit(str[i]) {
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
func matchArticleTitleByLength(content, article string) bool {
	// Normalize both the content and the article title
	normalizedContent := utils.NormalizeText(content)
	normalizedArticle := utils.NormalizeText(article)

	// Ensure the content is long enough to compare
	if len(normalizedContent) < len(normalizedArticle) {
		return false
	}

	// Compare the normalized article title with a substring of the normalized content
	return strings.HasPrefix(normalizedContent, normalizedArticle)
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
func findPageEndingWith(pdfPath, endsWith string, startPage, totalPages int, pageContents []string) (int, error) {
	for page := startPage; page <= totalPages; page++ {
		// Check if the pageContents slice has enough elements
		if page-1 >= len(pageContents) {
			return 0, fmt.Errorf("page %d exceeds the number of pages in the PDF", page)
		}

		// Get the normalized content for the current page
		normalizedContent := utils.NormalizeText(pageContents[page-1])
		logrus.Debugf("[DEBUG] Normalized content of page %d: '%s'", page, normalizedContent[:min(len(normalizedContent), 50)])

		// Check if the content ends with the specified text
		if strings.HasPrefix(normalizedContent, utils.NormalizeText(endsWith)) {
			logrus.Debugf("[DEBUG] Found 'ends-with' match on page %d: '%s'", page, normalizedContent[:min(len(normalizedContent), 50)])
			return page, nil
		}
	}
	return 0, nil
}
