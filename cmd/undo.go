package cmd

import (
	"pdf-extractor/internal/actions"

	"github.com/spf13/cobra"
)

var UndoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo the last action on generated PDF file",
	Long:  `The undo command restores the latest backup of the specified PDF file.`,
	RunE:  undo,
}

func init() {
	UndoCmd.Flags().StringVarP(&file, "file", "f", "", "Path to the PDF file to undo")
	UndoCmd.MarkFlagRequired("file")

	// Add --backup-path flag with default value "./backup"
	UndoCmd.Flags().StringVar(&backupPath, "backup-path", "./backup", "Path to the backup directory")

	rootCmd.AddCommand(UndoCmd)
}
func undo(cmd *cobra.Command, args []string) error {
	var cmds []actions.Command
	cmds = append(cmds, &actions.UndoSettings{
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
