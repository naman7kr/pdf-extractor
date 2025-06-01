package actions

import "pdf-extractor/internal/services"

type IndexSettings struct {
	File       string
	OutputPath string
}

func (s *IndexSettings) Execute() error {
	return services.ExtractIndex(s.File, s.OutputPath)
}

func (s *IndexSettings) Description() string {
	return "IndexExtractorCommand"
}
