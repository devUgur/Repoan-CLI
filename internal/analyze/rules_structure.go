package analyze

import (
	"context"

	"github.com/repoan/repoan/internal/model"
	"github.com/repoan/repoan/internal/structure"
)

type StructureAnalyzer struct {
	RepoRoot            string
	TopTokens           int
	IncludeTests        bool
	InstabilityHigh     float64
	ClusterCouplingHigh float64
	ClusterBoundaryLow  float64
}

func (s *StructureAnalyzer) Name() string {
	return "structure"
}

func (s *StructureAnalyzer) Analyze(ctx context.Context, snap *model.Snapshot) ([]model.Finding, error) {
	report, err := structure.Analyze(s.RepoRoot, snap, structure.Options{
		TopTokens:           s.TopTokens,
		IncludeTests:        s.IncludeTests,
		InstabilityHigh:     s.InstabilityHigh,
		ClusterCouplingHigh: s.ClusterCouplingHigh,
		ClusterBoundaryLow:  s.ClusterBoundaryLow,
	})
	if err != nil {
		return nil, err
	}
	return structure.FindingsFromReport(report, structure.Options{
		InstabilityHigh:     s.InstabilityHigh,
		ClusterCouplingHigh: s.ClusterCouplingHigh,
		ClusterBoundaryLow:  s.ClusterBoundaryLow,
	}), nil
}
