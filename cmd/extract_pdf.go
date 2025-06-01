package cmd

import (
	"pdf-extractor/internal/actions"

	"github.com/spf13/cobra"
)

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

	rootCmd.AddCommand(PDFExtractorCommand)
}
func extractPDF(cmd *cobra.Command, args []string) error {
	var cmds []actions.Command
	cmds = append(cmds, &actions.ExtractPDFSettings{
		File:       file,
		OutputPath: outputPath,
		ConfigPath: configPath,
		EndsWith:   endsWith,
	})
	invoker := actions.Invoker{
		Command: cmds,
	}
	if err := invoker.ExecuteCommand(); err != nil {
		return err
	}
	return nil
}
