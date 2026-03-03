package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"path/filepath"

	"github.com/repoan/repoan/internal/config"
	"github.com/repoan/repoan/internal/logging"
	"github.com/spf13/cobra"
)

var Version = "dev"
var Debug bool
var ConfigPath string
var Logger *slog.Logger
var Cfg *config.Config

var rootCmd = &cobra.Command{
	Use:           "repoan",
	Short:         "repoan is a repository analytics CLI",
	Long:          `repoan is a professional, fast, and cross-platform CLI tool 
designed to snapshot folder structures and provide repository insights.
It helps in documenting codebases and preparing repository context.`,
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		Logger = logging.New(Debug)
		
		// Load config
		var err error
		if ConfigPath != "" {
			Cfg, err = config.Load(ConfigPath)
			if err != nil {
				Logger.Warn("could not load specified config, using defaults", "path", ConfigPath, "err", err)
				Cfg = config.Default()
			}
		} else {
			// Try default path
			defaultPath := filepath.Join(".repoan", "config.yml")
			if _, err := os.Stat(defaultPath); err == nil {
				Cfg, err = config.Load(defaultPath)
				if err != nil {
					Logger.Warn("could not load default config, using defaults", "path", defaultPath, "err", err)
					Cfg = config.Default()
				}
			} else {
				Cfg = config.Default()
			}
		}
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
	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "", "Path to config file (default is .repoan/config.yml)")
}
