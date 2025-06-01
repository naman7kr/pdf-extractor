package cmd

import (
	"pdf-extractor/internal/actions"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete pdf file or pages in pdf file",
	Long:  `The delete command deletes a PDF file or specific pages in a PDF file`,
	RunE:  deletePDF,
}

func init() {
	DeleteCmd.Flags().StringVarP(&file, "file", "f", "", "Path to the PDF file")
	DeleteCmd.Flags().StringVar(&backupPath, "backup-path", "./backup", "Path to save backup files")
	DeleteCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(DeleteCmd)
}
func deletePDF(cmd *cobra.Command, args []string) error {
	var cmds []actions.Command
	cmds = append(cmds, &actions.DeleteSettings{
		File:       file,
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
