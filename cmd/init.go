package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/repoan/repoan/internal/config"
	"github.com/spf13/cobra"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize repoan in the current directory",
	Long: `Creates a .repoan/ directory with configuration and sets up .gitignore.
This is the professional way to manage repoan metadata in a repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repoanDir := ".repoan"
		dirs := []string{
			repoanDir,
			filepath.Join(repoanDir, "reports"),
			filepath.Join(repoanDir, "cache"),
			filepath.Join(repoanDir, "logs"),
		}

		for _, d := range dirs {
			if err := os.MkdirAll(d, 0755); err != nil {
				return fmt.Errorf("could not create directory %s: %w", d, err)
			}
			Logger.Debug("ensured directory exists", "dir", d)
		}

		// Create default config
		cfgPath := filepath.Join(repoanDir, "config.yml")
		if _, err := os.Stat(cfgPath); err == nil && !initForce {
			Logger.Info("config file already exists, skipping", "filename", cfgPath)
		} else {
			cfg := config.Default()
			if err := cfg.Save(cfgPath); err != nil {
				return fmt.Errorf("could not save default config: %w", err)
			}
			fmt.Printf("Created %s\n", cfgPath)
		}

		// Create baseline if not exists
		baselinePath := filepath.Join(repoanDir, "baseline.json")
		if _, err := os.Stat(baselinePath); os.IsNotExist(err) {
			if err := os.WriteFile(baselinePath, []byte("{}"), 0644); err != nil {
				return fmt.Errorf("could not create baseline.json: %w", err)
			}
			fmt.Printf("Created %s\n", baselinePath)
		}

		// Update .gitignore
		if err := updateGitignore(repoanDir); err != nil {
			return fmt.Errorf("could not update .gitignore: %w", err)
		}

		fmt.Println("Repoan initialized successfully in .repoan/")
		return nil
	},
}

func updateGitignore(repoanDir string) error {
	ignorePatterns := []string{
		fmt.Sprintf("/%s/cache/", repoanDir),
		fmt.Sprintf("/%s/logs/", repoanDir),
		fmt.Sprintf("/%s/reports/", repoanDir),
	}

	f, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	content, err := os.ReadFile(".gitignore")
	if err != nil {
		return err
	}

	existingContent := string(content)
	var toAdd []string
	for _, p := range ignorePatterns {
		if !strings.Contains(existingContent, p) {
			toAdd = append(toAdd, p)
		}
	}

	if len(toAdd) > 0 {
		if !strings.HasSuffix(existingContent, "\n") && len(existingContent) > 0 {
			if _, err := f.WriteString("\n"); err != nil {
				return err
			}
		}
		
		if !strings.Contains(existingContent, "# Repoan metadata") {
			if _, err := f.WriteString("\n# Repoan metadata\n"); err != nil {
				return err
			}
		}

		for _, p := range toAdd {
			if _, err := f.WriteString(p + "\n"); err != nil {
				return err
			}
		}
		fmt.Println("Updated .gitignore with repoan patterns")
	}

	return nil
}

func init() {
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Force overwrite existing config files")
	rootCmd.AddCommand(initCmd)
}
