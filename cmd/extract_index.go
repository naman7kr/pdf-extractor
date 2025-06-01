package cmd

import (
	"pdf-extractor/internal/actions"

	"github.com/spf13/cobra"
)

var indexExtractorCmd = &cobra.Command{
	Use:   "extract-index",
	Short: "Extract authors and titles from a PDF file",
	Long:  `The extract-index command extracts authors and titles from a PDF file to config.json file`,
	RunE:  extractIndex,
}

func init() {
	indexExtractorCmd.Flags().StringVarP(&file, "file", "f", "", "Path to the PDF file")
	indexExtractorCmd.MarkFlagRequired("file")
	indexExtractorCmd.Flags().StringVarP(&outputPath, "output-path", "o", "./", "Path to save the output files")
	rootCmd.AddCommand(indexExtractorCmd)
}
func extractIndex(cmd *cobra.Command, args []string) error {
	var cmds []actions.Command
	cmds = append(cmds, &actions.IndexSettings{
		File:       file,
		OutputPath: outputPath,
	})
	invoker := actions.Invoker{
		Command: cmds,
	}
	if err := invoker.ExecuteCommand(); err != nil {
		return err
	}
	return nil
}
