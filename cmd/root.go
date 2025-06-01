package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	versionFlag bool
	version     = "unknown" // Default version, can be set at build time
	debug       bool
)

// Root command definition
var rootCmd = &cobra.Command{
	Use:   "pdf-extractor",
	Short: "A simple tool to extract information from PDF-like files",
	Long: `pdf-extractor is a command-line tool that aims to extract
information from various document formats. Currently, it has a
subcommand to 'get chapters' from a DOCX file (though the actual
extraction is not yet implemented).`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
			logrus.Debug("Debug mode is enabled")
		} else {
			logrus.SetLevel(logrus.InfoLevel)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Handle the --version or -v flag
		if versionFlag {
			fmt.Println("PDF Extractor version:", version)
			return
		}

		// Default behavior if no flags or subcommands are provided
		fmt.Println("Please specify a subcommand or flag. Use --help for more information.")
	},
}

type Article struct {
	Title  string
	Author string
}

// Define the YAML structure
type ArticlesConfig struct {
	Articles []Article `yaml:"articles"`
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
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "Print the version number of pdf-extractor")

}
