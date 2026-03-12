package cmd

import (
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

var rulesOut string
var rulesFormat string
var rulesMaxDepth int
var rulesIgnore []string
var rulesRespectGitignore bool
var rulesHere bool

var rulesCmd = &cobra.Command{
    Use:   "rules [path]",
    Short: "Analyze repository boundary rules and clusters",
    Long:  `Scans the repository and computes structural clusters and boundary metrics (used as baseline for rules).`,
    RunE: func(cmd *cobra.Command, args []string) error {
        root := "."
        if len(args) > 0 {
            root = args[0]
        }

        if !rulesHere {
            gitRoot, err := git.FindRoot(root)
            if err == nil && gitRoot != "" {
                root = gitRoot
            }
        }

        // Merge config and flags
        ignorePatterns := config.EffectiveIgnorePatterns(Cfg.Scan.Ignore)
        if len(rulesIgnore) > 0 {
            ignorePatterns = config.EffectiveIgnorePatterns(rulesIgnore)
        }

        maxDepth := rulesMaxDepth
        if maxDepth == 0 {
            maxDepth = Cfg.Scan.MaxDepth
        }

        respectGitignore := rulesRespectGitignore
        if !cmd.Flags().Changed("respect-gitignore") {
            respectGitignore = Cfg.Scan.RespectGitignore
        }

        opts := scan.ScanOptions{
            Root:             root,
            IgnorePatterns:   ignorePatterns,
            MaxDepth:         maxDepth,
            RespectGitignore: respectGitignore,
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

        report, err := structure.Analyze(root, &snap, structure.Options{})
        if err != nil {
            return err
        }

        format := rulesFormat
        if !cmd.Flags().Changed("format") {
            format = Cfg.Output.DefaultFormat
        }

        formatted, err := output.FormatRulesReport(&model.RulesReport{
            GeneratedAt:      report.GeneratedAt,
            Root:             report.Root,
            ClusterMetrics:   func() []model.ClusterMetric { m := make([]model.ClusterMetric, 0, len(report.ClusterMetrics)); for _, c := range report.ClusterMetrics { m = append(m, model.ClusterMetric{Cluster: c.Cluster, Modules: c.Modules, InternalEdges: c.InternalEdges, ExternalInEdges: c.ExternalInEdges, ExternalOutEdges: c.ExternalOutEdges, CohesionScore: c.CohesionScore, CouplingScore: c.CouplingScore, BoundaryStrength: c.BoundaryStrength, DirectionalityScore: c.DirectionalityScore, Instability: c.Instability, InCycleModules: c.InCycleModules}) }; return m }(),
            StrongComponents: report.StrongComponents,
        }, format)
        if err != nil {
            return err
        }

        if rulesOut != "" {
            if err := os.WriteFile(rulesOut, []byte(formatted), 0644); err != nil {
                return fmt.Errorf("could not write output to %s: %w", rulesOut, err)
            }
            fmt.Printf("Rules report written to %s\n", rulesOut)
        } else {
            fmt.Println(formatted)
        }

        return nil
    },
}

func init() {
    rulesCmd.Flags().StringVarP(&rulesOut, "out", "o", "", "Output file path (e.g., rules.json)")
    rulesCmd.Flags().StringVarP(&rulesFormat, "format", "f", "", "Output format (text, md, json, mermaid)")
    rulesCmd.Flags().IntVarP(&rulesMaxDepth, "max-depth", "d", 0, "Maximum depth to scan")
    rulesCmd.Flags().StringSliceVarP(&rulesIgnore, "ignore", "i", []string{}, "Patterns to ignore")
    rulesCmd.Flags().BoolVar(&rulesRespectGitignore, "respect-gitignore", true, "Respect .gitignore files")
    rulesCmd.Flags().BoolVar(&rulesHere, "here", false, "Start scanning from current directory instead of git root")

    rootCmd.AddCommand(rulesCmd)
}
