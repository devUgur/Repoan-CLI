package analyze

import (
	"context"

	"github.com/repoan/repoan/internal/model"
)

// Analyzer defines the interface for all analysis rules.
type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, snap *model.Snapshot) ([]model.Finding, error)
}
