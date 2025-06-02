package actions

import "pdf-extractor/internal/services"

type DeleteSettings struct {
	File       string
	BackupPath string
	BackupFlag bool
}

func (s *DeleteSettings) Execute() error {
	return services.Delete(s.File, s.BackupPath, s.BackupFlag)
}

func (s *DeleteSettings) Description() string {
	return "DeleteCommand"
}
