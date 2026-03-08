package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/repoan/repoan/internal/config"
	"github.com/repoan/repoan/internal/git"
	"github.com/repoan/repoan/internal/model"
	"github.com/repoan/repoan/internal/scan"
	"github.com/spf13/cobra"
)

var snapshotOut string
var snapshotHere bool

var snapshotCmd = &cobra.Command{
	Use:   "snapshot [path]",
	Short: "Create a JSON snapshot of the repository",
	Long:  `Scans the directory structure and metadata, and exports it to a JSON file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		if !snapshotHere {
			gitRoot, err := git.FindRoot(root)
			if err == nil && gitRoot != "" {
				root = gitRoot
			}
		}

		opts := scan.ScanOptions{
			Root:             root,
			RespectGitignore: Cfg.Scan.RespectGitignore,
			IgnorePatterns:   config.EffectiveIgnorePatterns(Cfg.Scan.Ignore),
			MaxDepth:         Cfg.Scan.MaxDepth,
		}

		tree, err := scan.Scan(opts)
		if err != nil {
			return err
		}

		snap := model.Snapshot{
			RepoRoot:  root,
			Timestamp: time.Now(),
			Files:     []*model.FileItem{tree},
			Version:   Version,
		}

		data, err := json.MarshalIndent(snap, "", "  ")
		if err != nil {
			return err
		}

		if snapshotOut != "" {
			err := os.WriteFile(snapshotOut, []byte(data), 0644)
			if err != nil {
				return err
			}
			fmt.Printf("Snapshot written to %s\n", snapshotOut)
		} else {
			fmt.Println(string(data))
		}

		return nil
	},
}

func init() {
	snapshotCmd.Flags().StringVarP(&snapshotOut, "out", "o", "repoan.snapshot.json", "Output file path")
	snapshotCmd.Flags().BoolVar(&snapshotHere, "here", false, "Start scanning from current directory instead of git root")
	rootCmd.AddCommand(snapshotCmd)
}
