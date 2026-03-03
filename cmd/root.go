package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/repoan/repoan/internal/logging"
	"github.com/spf13/cobra"
)

var Version = "dev"
var Debug bool
var Logger *slog.Logger

var rootCmd = &cobra.Command{
	Use:   "repoan",
	Short: "repoan is a repository analytics CLI",
	Long: `repoan is a professional, fast, and cross-platform CLI tool 
designed to snapshot folder structures and provide repository insights.
It helps in documenting codebases and preparing repository context.`,
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		Logger = logging.New(Debug)
	},
}

func Execute(v string) {
	Version = v
	rootCmd.Version = Version
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(determineExitCode(err))
	}
}

func determineExitCode(err error) int {
	// Generic error handling
	return ExitGeneric
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&Debug, "debug", false, "Enable debug logging")
}
