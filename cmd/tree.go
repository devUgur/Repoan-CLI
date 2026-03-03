package cmd

import (
	"fmt"
	"os"

	"github.com/repoan/repoan/internal/git"
	"github.com/repoan/repoan/internal/output"
	"github.com/repoan/repoan/internal/scan"
	"github.com/spf13/cobra"
)

var treeOut string
var treeFormat string
var treeMaxDepth int
var treeIgnore []string
var treeRespectGitignore bool
var treeHere bool

var treeCmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Write the folder structure snapshot",
	Long:  `Scans the directory structure and writes it to a file or stdout. By default, it detects the git root.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		Logger.Debug("starting tree command", "root", root)

		if !treeHere {
			gitRoot, err := git.FindRoot(root)
			if err == nil && gitRoot != "" {
				Logger.Debug("detected git root", "gitRoot", gitRoot)
				root = gitRoot
			} else {
				Logger.Debug("no git root detected or error", "err", err)
			}
		}

		opts := scan.ScanOptions{
			Root:             root,
			IgnorePatterns:   treeIgnore,
			MaxDepth:         treeMaxDepth,
			RespectGitignore: treeRespectGitignore,
		}

		Logger.Debug("scanning directory", "options", opts)
		tree, err := scan.Scan(opts)
		if err != nil {
			return err
		}

		Logger.Debug("formatting output", "format", treeFormat)
		formatted := output.FormatTree(tree, treeFormat)

		if treeOut != "" {
			err := os.WriteFile(treeOut, []byte(formatted), 0644)
			if err != nil {
				return fmt.Errorf("could not write output to %s: %w", treeOut, err)
			}
			fmt.Printf("Tree structure written to %s\n", treeOut)
		} else {
			fmt.Println(formatted)
		}

		return nil
	},
}

func init() {
	treeCmd.Flags().StringVarP(&treeOut, "out", "o", "", "Output file path (e.g., structure.txt)")
	treeCmd.Flags().StringVarP(&treeFormat, "format", "f", "txt", "Output format (txt, md, json)")
	treeCmd.Flags().IntVarP(&treeMaxDepth, "max-depth", "d", 0, "Maximum depth to scan")
	treeCmd.Flags().StringSliceVarP(&treeIgnore, "ignore", "i", []string{".git", "node_modules", "dist", "build", ".venv", "vendor"}, "Patterns to ignore")
	treeCmd.Flags().BoolVar(&treeRespectGitignore, "respect-gitignore", true, "Respect .gitignore files")
	treeCmd.Flags().BoolVar(&treeHere, "here", false, "Start scanning from current directory instead of git root")

	rootCmd.AddCommand(treeCmd)
}
