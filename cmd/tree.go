package cmd

import (
	"fmt"
	"os"

	"github.com/repoan/repoan/internal/config"
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
var treeASCII bool
var treeDirsOnly bool

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

		// Merge config and flags
		ignorePatterns := config.EffectiveIgnorePatterns(Cfg.Scan.Ignore)
		if len(treeIgnore) > 0 {
			ignorePatterns = config.EffectiveIgnorePatterns(treeIgnore)
		}

		maxDepth := treeMaxDepth
		if maxDepth == 0 {
			maxDepth = Cfg.Scan.MaxDepth
		}

		respectGitignore := treeRespectGitignore
		// If flag was not explicitly set, use config
		if !cmd.Flags().Changed("respect-gitignore") {
			respectGitignore = Cfg.Scan.RespectGitignore
		}

		opts := scan.ScanOptions{
			Root:             root,
			IgnorePatterns:   ignorePatterns,
			MaxDepth:         maxDepth,
			RespectGitignore: respectGitignore,
			DirsOnly:         treeDirsOnly,
		}

		Logger.Debug("scanning directory", "options", opts)
		tree, err := scan.Scan(opts)
		if err != nil {
			return err
		}

		format := treeFormat
		if !cmd.Flags().Changed("format") {
			format = Cfg.Output.DefaultFormat
		}

		useASCII := treeASCII
		if !cmd.Flags().Changed("ascii") {
			// Redirects/pipes are typically consumed by tools that expect plain ASCII.
			if info, statErr := os.Stdout.Stat(); statErr == nil && (info.Mode()&os.ModeCharDevice) == 0 {
				useASCII = true
			}
		}

		Logger.Debug("formatting output", "format", format, "ascii", useASCII)
		formatted := output.FormatTree(tree, output.TreeOptions{
			Format: format,
			ASCII:  useASCII,
		})

		outputFile := treeOut
		if outputFile == "" {
			// If we want a default output file from config, we could add it here
		}

		if outputFile != "" {
			err := os.WriteFile(outputFile, []byte(formatted), 0644)
			if err != nil {
				return fmt.Errorf("could not write output to %s: %w", outputFile, err)
			}
			fmt.Printf("Tree structure written to %s\n", outputFile)
		} else {
			fmt.Println(formatted)
		}

		return nil
	},
}

func init() {
	treeCmd.Flags().StringVarP(&treeOut, "out", "o", "", "Output file path (e.g., structure.txt)")
	treeCmd.Flags().StringVarP(&treeFormat, "format", "f", "", "Output format (txt, md, json)")
	treeCmd.Flags().IntVarP(&treeMaxDepth, "max-depth", "d", 0, "Maximum depth to scan")
	treeCmd.Flags().StringSliceVarP(&treeIgnore, "ignore", "i", []string{}, "Patterns to ignore")
	treeCmd.Flags().BoolVar(&treeRespectGitignore, "respect-gitignore", true, "Respect .gitignore files")
	treeCmd.Flags().BoolVar(&treeHere, "here", false, "Start scanning from current directory instead of git root")
	treeCmd.Flags().BoolVar(&treeASCII, "ascii", false, "Use ASCII characters for tree (useful for file redirection)")
	treeCmd.Flags().BoolVar(&treeDirsOnly, "dirs-only", false, "Only show directories in tree")

	rootCmd.AddCommand(treeCmd)
}
