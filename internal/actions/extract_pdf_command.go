package actions

import "pdf-extractor/internal/services"

type ExtractPDFSettings struct {
	File       string
	OutputPath string
	ConfigPath string
	EndsWith   string
}

func (s *ExtractPDFSettings) Execute() error {
	return services.ExtractPDF(s.File, s.OutputPath, s.ConfigPath, s.EndsWith)
}

func (s *ExtractPDFSettings) Description() string {
	return "PDFExtractorCommand"
}
