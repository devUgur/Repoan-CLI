package analyze

import (
	"context"
	"github.com/repoan/repoan/internal/model"
)

// Severity represents the importance of a finding.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// Finding represents a single issue found by an analyzer.
type Finding struct {
	RuleID      string   `json:"rule_id"`
	Message     string   `json:"message"`
	Path        string   `json:"path"`
	Severity    Severity `json:"severity"`
	Suggestion  string   `json:"suggestion,omitempty"`
}

// Analyzer defines the interface for all analysis rules.
type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, snap *model.Snapshot) ([]Finding, error)
}
