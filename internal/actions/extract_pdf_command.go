package actions

import "pdf-extractor/internal/services"

type ExtractPDFSettings struct {
	File         string
	OutputPath   string
	ConfigPath   string
	EndsWith     string
	FromPage     int
	ToPage       int
	ArticleTitle string
}

func (s *ExtractPDFSettings) Execute() error {
	if s.FromPage != -1 || s.ToPage != -1 {
		return services.ExtractPDFFromRange(s.File, s.OutputPath, s.FromPage, s.ToPage, s.ArticleTitle)
	}
	return services.ExtractPDF(s.File, s.OutputPath, s.ConfigPath, s.EndsWith)
}

func (s *ExtractPDFSettings) Description() string {
	return "PDFExtractorCommand"
}
