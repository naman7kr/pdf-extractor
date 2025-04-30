package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pdf-extractor",
	Short: "A simple tool to extract information from PDF-like files",
	Long: `pdf-extractor is a command-line tool that aims to extract
information from various document formats. Currently, it has a
subcommand to 'get chapters' from a DOCX file (though the actual
extraction is not yet implemented).`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you can define global flags and other configurations
	// that apply to all commands.
}