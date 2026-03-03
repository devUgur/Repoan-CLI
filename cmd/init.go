package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize repoan in the current directory",
	Long:  `Creates default configuration files like .repoan.yml and .repoanignore.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files := map[string]string{
			".repoan.yml":   "ignore_patterns:\n  - .git\n  - node_modules\n  - dist\n  - build\n  - .venv\n  - vendor\n",
			".repoanignore": ".git\nnode_modules\ndist\nbuild\n.venv\nvendor\n.DS_Store\n",
		}

		for filename, content := range files {
			if _, err := os.Stat(filename); err == nil && !initForce {
				Logger.Info("file already exists, skipping", "filename", filename)
				continue
			}

			err := os.WriteFile(filename, []byte(content), 0644)
			if err != nil {
				return fmt.Errorf("could not create %s: %w", filename, err)
			}
			fmt.Printf("Created %s\n", filename)
		}

		fmt.Println("Repoan initialized successfully!")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Force overwrite existing config files")
	rootCmd.AddCommand(initCmd)
}
