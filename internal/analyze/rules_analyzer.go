package analyze

import (
    "context"

    "github.com/repoan/repoan/internal/model"
    "github.com/repoan/repoan/internal/structure"
)

// ComputeRulesReport runs the structural analysis and converts the relevant
// cluster/boundary information into a `model.RulesReport`.
func ComputeRulesReport(ctx context.Context, repoRoot string, snap *model.Snapshot, opts structure.Options) (*model.RulesReport, error) {
    report, err := structure.Analyze(repoRoot, snap, opts)
    if err != nil {
        return nil, err
    }

    out := &model.RulesReport{
        GeneratedAt:    report.GeneratedAt,
        Root:           report.Root,
        ClusterMetrics: make([]model.ClusterMetric, 0, len(report.ClusterMetrics)),
        StrongComponents: report.StrongComponents,
    }

    for _, c := range report.ClusterMetrics {
        out.ClusterMetrics = append(out.ClusterMetrics, model.ClusterMetric{
            Cluster:             c.Cluster,
            Modules:             c.Modules,
            InternalEdges:       c.InternalEdges,
            ExternalInEdges:     c.ExternalInEdges,
            ExternalOutEdges:    c.ExternalOutEdges,
            CohesionScore:       c.CohesionScore,
            CouplingScore:       c.CouplingScore,
            BoundaryStrength:    c.BoundaryStrength,
            DirectionalityScore: c.DirectionalityScore,
            Instability:         c.Instability,
            InCycleModules:      c.InCycleModules,
        })
    }

    return out, nil
}
