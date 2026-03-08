package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/repoan/repoan/internal/config"
	"github.com/repoan/repoan/internal/git"
	"github.com/repoan/repoan/internal/model"
	"github.com/repoan/repoan/internal/output"
	"github.com/repoan/repoan/internal/scan"
	"github.com/repoan/repoan/internal/structure"
	"github.com/spf13/cobra"
)

var structureOut string
var structureFormat string
var structureTopTokens int
var structureIncludeTests bool
var structureHere bool
var structureRawDependencies bool
var structureRawDependenciesLimit int
var structureRawDependenciesOnlyUnresolved bool
var structureRawDependenciesOnlyExternal bool
var structureRawDependenciesMinConfidence float64

var structureCmd = &cobra.Command{
	Use:   "structure [path]",
	Short: "Analyze repository structure and dependency shape",
	Long:  `Computes structural token statistics, module dependency metrics, SCC cycles, and layer candidates.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		if !structureHere {
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

		snap := &model.Snapshot{
			RepoRoot:  root,
			Timestamp: time.Now(),
			Files:     []*model.FileItem{tree},
			Version:   Version,
		}

		topTokens := structureTopTokens
		if !cmd.Flags().Changed("top-tokens") {
			topTokens = Cfg.Analysis.Structure.TopTokens
		}

		includeTests := structureIncludeTests
		if !cmd.Flags().Changed("include-tests") {
			includeTests = Cfg.Analysis.Structure.IncludeTests
		}

		includeRawDependencies := structureRawDependencies
		rawDependenciesLimit := structureRawDependenciesLimit
		rawDependenciesOnlyUnresolved := structureRawDependenciesOnlyUnresolved
		rawDependenciesOnlyExternal := structureRawDependenciesOnlyExternal
		rawDependenciesMinConfidence := structureRawDependenciesMinConfidence

		if rawDependenciesMinConfidence < 0 || rawDependenciesMinConfidence > 1 {
			return fmt.Errorf("--raw-dependencies-min-confidence must be between 0 and 1")
		}

		report, err := structure.Analyze(root, snap, structure.Options{
			TopTokens:           topTokens,
			IncludeTests:        includeTests,
			InstabilityHigh:     Cfg.Analysis.Structure.InstabilityHigh,
			ClusterCouplingHigh: Cfg.Analysis.Structure.ClusterCouplingHigh,
			ClusterBoundaryLow:  Cfg.Analysis.Structure.ClusterBoundaryLow,
			RawDependencies: structure.RawDependencyOptions{
				Enabled:        includeRawDependencies,
				Limit:          rawDependenciesLimit,
				OnlyUnresolved: rawDependenciesOnlyUnresolved,
				OnlyExternal:   rawDependenciesOnlyExternal,
				MinConfidence:  rawDependenciesMinConfidence,
			},
		})
		if err != nil {
			return err
		}

		format := structureFormat
		if !cmd.Flags().Changed("format") {
			format = Cfg.Output.DefaultFormat
		}
		if format == "" {
			format = "text"
		}

		var rendered string
		switch format {
		case "sarif":
			findings := structure.FindingsFromReport(report, structure.Options{
				InstabilityHigh:     Cfg.Analysis.Structure.InstabilityHigh,
				ClusterCouplingHigh: Cfg.Analysis.Structure.ClusterCouplingHigh,
				ClusterBoundaryLow:  Cfg.Analysis.Structure.ClusterBoundaryLow,
			})
			sarif, err := output.FormatSarif(findings, "repoan-structure", Version)
			if err != nil {
				return err
			}
			rendered = sarif
		case "json":
			data, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return err
			}
			rendered = string(data)
		default:
			rendered, err = output.FormatStructureReport(report, format)
			if err != nil {
				return err
			}
		}

		if structureOut != "" {
			if err := os.WriteFile(structureOut, []byte(rendered), 0644); err != nil {
				return err
			}
			fmt.Printf("Structure report written to %s\n", structureOut)
			return nil
		}

		fmt.Println(rendered)
		return nil
	},
}

func init() {
	structureCmd.Flags().StringVarP(&structureOut, "out", "o", "", "Output file path")
	structureCmd.Flags().StringVarP(&structureFormat, "format", "f", "text", "Output format (text, md, json, sarif, mermaid, mermaid-cluster)")
	structureCmd.Flags().IntVar(&structureTopTokens, "top-tokens", 25, "Number of top structural tokens to include")
	structureCmd.Flags().BoolVar(&structureIncludeTests, "include-tests", false, "Include test files in dependency calculations")
	structureCmd.Flags().BoolVar(&structureRawDependencies, "raw-dependencies", false, "Include raw file-level dependency trace data (debug)")
	structureCmd.Flags().IntVar(&structureRawDependenciesLimit, "raw-dependencies-limit", 0, "Limit number of raw dependency records (0 = unlimited)")
	structureCmd.Flags().BoolVar(&structureRawDependenciesOnlyUnresolved, "raw-dependencies-only-unresolved", false, "Only include unresolved entries in raw dependency trace output")
	structureCmd.Flags().BoolVar(&structureRawDependenciesOnlyExternal, "raw-dependencies-only-external", false, "Only include external entries in raw dependency trace output")
	structureCmd.Flags().Float64Var(&structureRawDependenciesMinConfidence, "raw-dependencies-min-confidence", 0, "Only include raw dependency records with confidence >= threshold (0..1)")
	structureCmd.Flags().BoolVar(&structureHere, "here", false, "Start scanning from current directory instead of git root")
	rootCmd.AddCommand(structureCmd)
}
