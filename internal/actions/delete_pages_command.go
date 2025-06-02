package actions

import "pdf-extractor/internal/services"

type DeletePagesSettings struct {
	File       string
	FromPage   int
	ToPage     int
	AtPage     int
	StartsWith string
	BackupPath string
	BackupFlag bool
}

func (s *DeletePagesSettings) Execute() error {
	return services.DeletePages(s.File, s.FromPage, s.ToPage, s.AtPage, s.StartsWith, s.BackupPath, s.BackupFlag)
}

func (s *DeletePagesSettings) Description() string {
	return "DeletePagesCommand"
}
