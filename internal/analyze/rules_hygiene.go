package analyze

import (
	"context"
	"fmt"

	"github.com/repoan/repoan/internal/model"
)

type HygieneAnalyzer struct {
	MaxFileSize int64
}

func (h *HygieneAnalyzer) Name() string {
	return "hygiene"
}

func (h *HygieneAnalyzer) Analyze(ctx context.Context, snap *model.Snapshot) ([]Finding, error) {
	var findings []Finding

	for _, file := range snap.Files {
		findings = append(findings, h.analyzeRecursive(file)...)
	}

	return findings, nil
}

func (h *HygieneAnalyzer) analyzeRecursive(file *model.FileItem) []Finding {
	var findings []Finding

	// Large file check
	if !file.IsDir && file.Size > h.MaxFileSize {
		findings = append(findings, Finding{
			RuleID:     "REP-HYG-001",
			Message:    fmt.Sprintf("Large file detected (%d bytes)", file.Size),
			Path:       file.Path,
			Severity:   SeverityWarning,
			Suggestion: "Consider using Git LFS or removing large binaries from the repository.",
		})
	}

	for _, child := range file.Children {
		findings = append(findings, h.analyzeRecursive(child)...)
	}

	return findings
}
