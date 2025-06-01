package services

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"pdf-extractor/internal/models"
	"pdf-extractor/internal/utils"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func ExtractIndex(file string, outputPath string) error {
	// Ensure the output directory exists
	err := utils.CreateDirectoryIfNotExists(outputPath)
	if err != nil {
		return err
	}

	// Extract the page with the title "Contents" directly into memory
	contentsPage, err := extractContentsPageInMemory(file)
	if err != nil {
		return fmt.Errorf("error extracting content: %v", err)
	}

	// Parse the extracted content to extract titles and authors
	articles, err := parseTitlesAndAuthorsFromContent(contentsPage)
	if err != nil {
		return fmt.Errorf("error parsing titles and authors: %v", err)
	}

	// Debug: Print the articles array
	logrus.Debugf("Extracted Articles: %+v\n", articles)

	// Save articles and authors to config.yaml
	yamlFilePath := filepath.Join(outputPath, "config.yaml")
	err = saveArticlesAndAuthorsToYAML(articles, yamlFilePath)
	if err != nil {
		return fmt.Errorf("error saving articles and authors to yaml: %v", err)
	}

	logrus.Infof("Articles and authors saved successfully to %s", yamlFilePath)
	return nil
}
func extractContentsPageInMemory(pdfPath string) (string, error) {
	// Get the total number of pages in the PDF
	totalPages, err := utils.GetPDFPageCount(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to get page count: %v", err)
	}

	// Iterate through each page to find the "Contents" page
	for page := 1; page <= totalPages; page++ {
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("page_%d.txt", page))

		// Extract the current page using pdftotext
		err := utils.ExtractPDFPageWithPdftotext(pdfPath, tempFile, page)
		if err != nil {
			return "", fmt.Errorf("failed to extract page %d: %v", page, err)
		}

		// Read the extracted content from the temporary file
		content, err := os.ReadFile(tempFile)
		if err != nil {
			return "", fmt.Errorf("failed to read temporary file for page %d: %v", page, err)
		}
		// Check if the page contains the word "Contents"
		// also normalize before compare case

		if strings.Contains(utils.NormalizeText(string(content)), "contents") {
			// Remove the temporary file
			os.Remove(tempFile)

			fmt.Printf("Found 'Contents' on page %d\n", page)
			return string(content), nil
		}

		// Remove the temporary file
		os.Remove(tempFile)
	}

	return "", fmt.Errorf("'Contents' page not found in the PDF")
}
func parseTitlesAndAuthorsFromContent(content string) ([]models.Article, error) {
	// Regular expression to match article numbers (e.g., "1.")
	numberRegex := regexp.MustCompile(`^\d+\.\s*`)
	// Regular expression to match lines with only a number or a number range (e.g., "1", "10", "1-10")
	numberOrRangeRegex := regexp.MustCompile(`^\d+(-\d+)?$`)

	var articles []models.Article
	scanner := bufio.NewScanner(strings.NewReader(content))
	var titleLines []string
	var prevText string     // To store the last appended text
	var foundIndex bool     // Flag to indicate if an article index has been found
	var expectingTitle bool // Flag to indicate if the next line is the title

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fmt.Printf("Scanning line: '%s'\n", line) // Debugging line

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip lines that contain only a number or a number range
		if numberOrRangeRegex.MatchString(line) {
			fmt.Printf("Skipping line with only a number or range: '%s'\n", line) // Debugging line
			continue
		}

		// Match article numbers
		if numberRegex.MatchString(line) {
			fmt.Printf("Detected article number: '%s'\n", line) // Debugging line
			foundIndex = true

			// If a new article number is found, save the previous article (if any)
			if len(titleLines) > 0 {
				// Remove the last line from titleLines (it is the author)
				if len(titleLines) > 1 {
					prevText = titleLines[len(titleLines)-1]
					titleLines = titleLines[:len(titleLines)-1]
				} else if len(titleLines) == 1 {
					// If there's only one line, treat it as the title and leave prevText empty
					prevText = ""
				}

				// Combine all title lines into a single title
				for i := range titleLines {
					titleLines[i] = utils.TrimTrailingNumber(titleLines[i])
				}
				title := strings.Join(titleLines, " ")
				title = strings.TrimSpace(title)
				title = strings.TrimSuffix(title, ".")
				// trim any trailing number or digits
				title = utils.TrimTrailingNumber(title)
				articles = append(articles, models.Article{
					Title:  title,
					Author: prevText,
				})

				fmt.Printf("Added article: Title='%s', Author='%s'\n", title, prevText) // Debugging line
				titleLines = nil                                                        // Reset for the next article
			}

			// Extract the title from the same line if it contains both the index and the title
			title := numberRegex.ReplaceAllString(line, "")
			if title != "" {
				titleLines = append(titleLines, title)
				expectingTitle = false // Reset the flag since the title is already extracted
			} else {
				expectingTitle = true // The next line(s) should be treated as the title
			}
		} else if expectingTitle {
			// Append the current line to the title
			fmt.Printf("Appending to title: '%s'\n", line) // Debugging line
			titleLines = append(titleLines, line)
			expectingTitle = false // Reset the flag after appending the first title line
		} else if foundIndex {
			// Append the current line to the title if an index has been found
			fmt.Printf("Appending to title: '%s'\n", line) // Debugging line
			titleLines = append(titleLines, line)
		}
	}

	// Handle the last article
	if len(titleLines) > 0 {
		// Remove the last line from titleLines (it is the author)
		if len(titleLines) > 1 {
			prevText = titleLines[len(titleLines)-1]
			titleLines = titleLines[:len(titleLines)-1]
		} else if len(titleLines) == 1 {
			// If there's only one line, treat it as the title and leave prevText empty
			prevText = ""
		}
		for i := range titleLines {
			titleLines[i] = utils.TrimTrailingNumber(titleLines[i])
		}
		// Combine all title lines into a single title
		title := strings.Join(titleLines, " ")
		title = strings.TrimSpace(title)
		title = strings.TrimSuffix(title, ".")
		title = utils.TrimTrailingNumber(title)
		articles = append(articles, models.Article{
			Title:  title,
			Author: prevText,
		})
		fmt.Printf("Added last article: Title='%s', Author='%s'\n", title, prevText) // Debugging line
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading content: %v", err)
	}

	return articles, nil
}
func saveArticlesAndAuthorsToYAML(articles []models.Article, filePath string) error {
	// Create the YAML structure
	config := models.ArticlesConfig{
		Articles: articles,
	}

	// Create or overwrite the YAML file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create YAML file: %v", err)
	}
	defer file.Close()

	// Encode the structure into YAML and write to the file
	encoder := yaml.NewEncoder(file)
	err = encoder.Encode(config)
	if err != nil {
		return fmt.Errorf("failed to write YAML data: %v", err)
	}

	fmt.Printf("Saved articles and authors to YAML file: %s\n", filePath)
	return nil
}
