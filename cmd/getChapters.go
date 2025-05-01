package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var file string

type Article struct {
	Title  string
	Author string
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get information from a file",
	Long:  `Allows you to retrieve specific information from a given file.`,
}
var chaptersCmd = &cobra.Command{
	Use:     "chapters",
	Short:   "Read all pages of a DOCX file",
	Aliases: []string{"pages"},
	Long:    `Reads all the pages of a Microsoft Word DOCX file.`,
	Run:     processChapters,
}

func init() {
	// Add 'get' as a subcommand of the root command
	rootCmd.AddCommand(getCmd)

	// Add 'chapters' as a subcommand of 'get'
	getCmd.AddCommand(chaptersCmd)

	// Define and mark the --file flag for the 'chapters' command
	chaptersCmd.Flags().StringVarP(&file, "file", "f", "", "Path to the DOCX file")
	chaptersCmd.MarkFlagRequired("file")
}

func processChapters(cmd *cobra.Command, args []string) {
	fmt.Printf("Processing file: %s to find the 'Contents' page...\n", file)

	// Extract the page with the title "Contents" directly into memory
	contentsPage, err := extractContentsPageInMemory(file)
	if err != nil {
		fmt.Printf("Error extracting content: %v\n", err)
		os.Exit(1)
	}

	// Parse the extracted content to extract titles and authors
	articles, err := parseTitlesAndAuthorsFromContent(contentsPage)
	if err != nil {
		fmt.Printf("Error parsing titles and authors: %v\n", err)
		os.Exit(1)
	}

	// Debug: Print the articles array
	fmt.Printf("Extracted Articles: %+v\n", articles)

	// Save the article titles to articles.txt
	err = saveArticleTitles(articles, "articles.txt")
	if err != nil {
		fmt.Printf("Error saving article titles: %v\n", err)
		os.Exit(1)
	}
	// Save authors to authors.txt
	err = saveAuthors(articles, "authors.txt")
	if err != nil {
		fmt.Printf("Error saving authors: %v\n", err)
		return
	}
	fmt.Println("Titles and authors saved successfully.")
}

func extractContentsPageInMemory(pdfPath string) (string, error) {
	// Get the total number of pages in the PDF
	totalPages, err := getPDFPageCount(pdfPath)
	if err != nil {
		return "", fmt.Errorf("failed to get page count: %v", err)
	}

	// Iterate through each page to find the "Contents" page
	for page := 1; page <= totalPages; page++ {
		tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("page_%d.txt", page))

		// Extract the current page using pdftotext
		err := extractPDFPageWithPdftotext(pdfPath, tempFile, page)
		if err != nil {
			return "", fmt.Errorf("failed to extract page %d: %v", page, err)
		}

		// Read the extracted content from the temporary file
		content, err := os.ReadFile(tempFile)
		if err != nil {
			return "", fmt.Errorf("failed to read temporary file for page %d: %v", page, err)
		}

		// Check if the page contains the word "Contents"
		if strings.Contains(string(content), "Contents") {
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

func parseTitlesAndAuthorsFromContent(content string) ([]Article, error) {
	// Regular expression to match article numbers (e.g., "1.")
	numberRegex := regexp.MustCompile(`^\d+\.\s*`)
	// Regular expression to match lines with only a number or a number range (e.g., "1", "10", "1-10")
	numberOrRangeRegex := regexp.MustCompile(`^\d+(-\d+)?$`)

	var articles []Article
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
				title := strings.Join(titleLines, " ")
				title = strings.TrimSpace(title)
				title = strings.TrimSuffix(title, ".")
				articles = append(articles, Article{
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

		// Combine all title lines into a single title
		title := strings.Join(titleLines, " ")
		title = strings.TrimSpace(title)
		title = strings.TrimSuffix(title, ".")
		articles = append(articles, Article{
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
func saveArticleTitles(articles []Article, filePath string) error {
	// Remove the existing file if it exists
	if _, err := os.Stat(filePath); err == nil {
		err = os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("failed to remove existing file: %v", err)
		}
	}

	// Create or overwrite the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write each article title to the file
	for _, article := range articles {
		_, err := file.WriteString(article.Title + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
		fmt.Printf("Saved article title: %s\n", article.Title) // Debugging line
	}

	return nil
}
func saveAuthors(articles []Article, filePath string) error {
	// Remove the existing file if it exists
	if _, err := os.Stat(filePath); err == nil {
		err = os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("failed to remove existing file: %v", err)
		}
	}

	// Create or overwrite the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write each author to the file
	for _, article := range articles {
		_, err := file.WriteString(article.Author + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to file: %v", err)
		}
		fmt.Printf("Saved author: %s\n", article.Author) // Debugging line
	}

	return nil
}
