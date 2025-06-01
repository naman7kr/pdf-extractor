package cmd

import (
	"pdf-extractor/internal/actions"

	"github.com/spf13/cobra"
)

var (
	fromPage   int
	toPage     int
	atPage     int
	startsWith string
)
var DeletePagesCommand = &cobra.Command{
	Use:   "delete-pages",
	Short: "Delete specific pages from a PDF file",
	Long:  `The pages command deletes specific pages from a PDF file based on the provided page numbers`,
	RunE:  deletePages,
}

func init() {
	DeletePagesCommand.Flags().StringVarP(&file, "file", "f", "", "Path to the PDF file")
	DeletePagesCommand.Flags().IntVar(&fromPage, "from", 0, "Starting page number to delete")
	DeletePagesCommand.Flags().IntVar(&toPage, "to", 0, "Ending page number to delete")
	DeletePagesCommand.Flags().IntVar(&atPage, "at", 0, "Specific page number to delete")
	DeletePagesCommand.Flags().StringVar(&startsWith, "starts-with", "", "Delete pages where content starts with the specified string")
	DeletePagesCommand.Flags().StringVar(&backupPath, "backup-path", "./backup", "Path to save backup files")
	DeletePagesCommand.MarkFlagRequired("file")
	rootCmd.AddCommand(DeletePagesCommand)
}
func deletePages(cmd *cobra.Command, args []string) error {
	var cmds []actions.Command
	cmds = append(cmds, &actions.DeletePagesSettings{
		File:       file,
		FromPage:   fromPage,
		ToPage:     toPage,
		AtPage:     atPage,
		StartsWith: startsWith,
		BackupPath: backupPath,
	})
	invoker := actions.Invoker{
		Command: cmds,
	}
	if err := invoker.ExecuteCommand(); err != nil {
		return err
	}
	return nil
}
