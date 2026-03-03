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

func (h *HygieneAnalyzer) Analyze(ctx context.Context, snap *model.Snapshot) ([]model.Finding, error) {
	var findings []model.Finding

	for _, file := range snap.Files {
		findings = append(findings, h.analyzeRecursive(file)...)
	}

	return findings, nil
}

func (h *HygieneAnalyzer) analyzeRecursive(file *model.FileItem) []model.Finding {
	var findings []model.Finding

	// Large file check
	if !file.IsDir && file.Size > h.MaxFileSize {
		finding := model.Finding{
			RuleID:      "REP-HYG-001",
			Title:       "Large File Detected",
			Message:     fmt.Sprintf("A large file was detected (%d bytes).", file.Size),
			Path:        file.Path,
			Severity:    model.SevMedium,
			Category:    model.CatHygiene,
			Remediation: "Consider using Git LFS or removing large binaries from the repository.",
		}
		finding.Fingerprint = model.FingerprintStable(finding.RuleID, finding.Path, 0, "large_file")
		findings = append(findings, finding)
	}

	for _, child := range file.Children {
		findings = append(findings, h.analyzeRecursive(child)...)
	}

	return findings
}
