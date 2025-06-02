package cmd

import (
	"fmt"
	"pdf-extractor/internal/actions"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var articleTitle string
var PDFExtractorCommand = &cobra.Command{
	Use:   "extract",
	Short: "Extract authors and titles from a PDF file",
	Long:  `The extract-index command extracts authors and titles from a PDF file to config.json file`,
	RunE:  extractPDF,
}

func init() {
	// Add --file flag with required validation
	PDFExtractorCommand.Flags().StringVarP(&file, "file", "f", "", "Path to the PDF file")
	PDFExtractorCommand.MarkFlagRequired("file")

	// Add --output-path flag with default value "extracted/"
	PDFExtractorCommand.Flags().StringVarP(&outputPath, "output-path", "o", "./extracted", "Path where the PDF files are generated")

	// Update --config-path flag to point to a directory
	PDFExtractorCommand.Flags().StringVarP(&configPath, "config-path", "c", "./configs", "Path to the directory containing the config.json file")

	// Add --ends-with flag
	PDFExtractorCommand.Flags().StringVar(&endsWith, "ends-with", "", "Text to find the page where the last article ends")
	// add from and to flags without default values

	PDFExtractorCommand.Flags().IntVar(&fromPage, "from", -1, "Starting page number to extract from")
	PDFExtractorCommand.Flags().IntVar(&toPage, "to", -1, "Ending page number to extract to")
	PDFExtractorCommand.Flags().StringVar(&articleTitle, "article-title", "", "Name of the article")

	rootCmd.AddCommand(PDFExtractorCommand)
}
func extractPDF(cmd *cobra.Command, args []string) error {
	var cmds []actions.Command
	if fromPage != -1 || toPage != -1 {
		if endsWith != "" {
			logrus.Warn("ends-with flag is redundant when using from and to flags. It will be ignored.")
		}
		if articleTitle == "" {
			return fmt.Errorf("article-title flag is required when using from and to flags")
		}
	}
	cmds = append(cmds, &actions.ExtractPDFSettings{
		File:         file,
		OutputPath:   outputPath,
		ConfigPath:   configPath,
		EndsWith:     endsWith,
		FromPage:     fromPage,
		ToPage:       toPage,
		ArticleTitle: articleTitle,
	})
	invoker := actions.Invoker{
		Command: cmds,
	}
	if err := invoker.ExecuteCommand(); err != nil {
		return err
	}
	return nil
}
