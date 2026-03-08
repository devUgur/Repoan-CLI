package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/repoan/repoan/internal/analyze"
	"github.com/repoan/repoan/internal/config"
	"github.com/repoan/repoan/internal/git"
	"github.com/repoan/repoan/internal/model"
	"github.com/repoan/repoan/internal/output"
	"github.com/repoan/repoan/internal/scan"
	"github.com/spf13/cobra"
)

var analyzeOut string
var analyzeFormat string
var analyzeFailOn string

var analyzeCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "Run analysis rules on the repository",
	Long:  `Scans the directory structure and runs security and hygiene checks.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}

		gitRoot, err := git.FindRoot(root)
		if err == nil && gitRoot != "" {
			root = gitRoot
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

		// Determine which analyzers to run
		var analyzers []analyze.Analyzer
		enabledRules := Cfg.Analysis.EnabledRules

		for _, rule := range enabledRules {
			switch rule {
			case "security":
				analyzers = append(analyzers, &analyze.SecurityAnalyzer{})
			case "large-files", "hygiene":
				analyzers = append(analyzers, &analyze.HygieneAnalyzer{MaxFileSize: 1024 * 1024 * 5}) // 5MB
			case "structure":
				analyzers = append(analyzers, &analyze.StructureAnalyzer{
					RepoRoot:            root,
					TopTokens:           Cfg.Analysis.Structure.TopTokens,
					IncludeTests:        Cfg.Analysis.Structure.IncludeTests,
					InstabilityHigh:     Cfg.Analysis.Structure.InstabilityHigh,
					ClusterCouplingHigh: Cfg.Analysis.Structure.ClusterCouplingHigh,
					ClusterBoundaryLow:  Cfg.Analysis.Structure.ClusterBoundaryLow,
				})
			}
		}

		// If no rules enabled in config, use defaults
		if len(analyzers) == 0 {
			analyzers = []analyze.Analyzer{
				&analyze.SecurityAnalyzer{},
				&analyze.HygieneAnalyzer{MaxFileSize: 1024 * 1024 * 5},
				&analyze.StructureAnalyzer{
					RepoRoot:            root,
					TopTokens:           Cfg.Analysis.Structure.TopTokens,
					IncludeTests:        Cfg.Analysis.Structure.IncludeTests,
					InstabilityHigh:     Cfg.Analysis.Structure.InstabilityHigh,
					ClusterCouplingHigh: Cfg.Analysis.Structure.ClusterCouplingHigh,
					ClusterBoundaryLow:  Cfg.Analysis.Structure.ClusterBoundaryLow,
				},
			}
		}

		allFindings := make([]model.Finding, 0)
		var findingsMutex sync.Mutex
		var wg sync.WaitGroup
		var errs []error
		var errMutex sync.Mutex

		for _, a := range analyzers {
			wg.Add(1)
			go func(analyzer analyze.Analyzer) {
				defer wg.Done()
				findings, err := analyzer.Analyze(context.Background(), snap)
				if err != nil {
					errMutex.Lock()
					errs = append(errs, err)
					errMutex.Unlock()
					return
				}
				findingsMutex.Lock()
				allFindings = append(allFindings, findings...)
				findingsMutex.Unlock()
			}(a)
		}
		wg.Wait()

		if len(errs) > 0 {
			return errs[0]
		}

		// Sort findings for determinism
		sort.Slice(allFindings, func(i, j int) bool {
			if allFindings[i].Path != allFindings[j].Path {
				return allFindings[i].Path < allFindings[j].Path
			}
			return allFindings[i].RuleID < allFindings[j].RuleID
		})

		format := analyzeFormat
		if !cmd.Flags().Changed("format") {
			format = Cfg.Output.DefaultFormat
		}

		var formatted string
		switch format {
		case "json":
			data, err := json.MarshalIndent(allFindings, "", "  ")
			if err != nil {
				return err
			}
			formatted = string(data)
		case "sarif":
			sarif, err := output.FormatSarif(allFindings, "repoan", Version)
			if err != nil {
				return err
			}
			formatted = sarif
		case "md", "text":
			// Human-friendly text output
			if len(allFindings) == 0 {
				fmt.Println("No findings found! ✨")
				return nil
			}
			for _, f := range allFindings {
				fmt.Printf("[%s] %s: %s\n  Path: %s\n  Suggestion: %s\n\n", f.Severity, f.RuleID, f.Message, f.Path, f.Remediation)
			}
			return nil
		}

		if analyzeOut != "" {
			err := os.WriteFile(analyzeOut, []byte(formatted), 0644)
			if err != nil {
				return err
			}
			fmt.Printf("Analysis results written to %s\n", analyzeOut)
		} else {
			fmt.Println(formatted)
		}

		// Fail on check
		failOn := analyzeFailOn
		if failOn == "" {
			failOn = Cfg.Analysis.FailOn
		}

		if failOn != "" && failOn != "none" {
			for _, f := range allFindings {
				if string(f.Severity) == failOn {
					fmt.Printf("Failing because a finding with severity '%s' was detected.\n", failOn)
					os.Exit(ExitFindings)
				}
			}
		}

		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeOut, "out", "o", "", "Output file path")
	analyzeCmd.Flags().StringVarP(&analyzeFormat, "format", "f", "text", "Output format (text, json, sarif)")
	analyzeCmd.Flags().StringVar(&analyzeFailOn, "fail-on", "", "Fail on specific severity (info, warning, high, critical)")
	rootCmd.AddCommand(analyzeCmd)
}
