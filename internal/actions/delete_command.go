package actions

import "pdf-extractor/internal/services"

type DeleteSettings struct {
	File       string
	BackupPath string
}

func (s *DeleteSettings) Execute() error {
	return services.Delete(s.File, s.BackupPath)
}

func (s *DeleteSettings) Description() string {
	return "DeleteCommand"
}
