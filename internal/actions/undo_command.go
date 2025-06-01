package actions

import "pdf-extractor/internal/services"

type UndoSettings struct {
	File       string
	BackupPath string
}

func (s *UndoSettings) Execute() error {
	return services.Undo(s.File, s.BackupPath)
}

func (s *UndoSettings) Description() string {
	return "UndoCommand"
}
