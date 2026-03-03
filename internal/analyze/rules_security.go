package analyze

import (
	"context"
	"strings"

	"github.com/repoan/repoan/internal/model"
)

type SecurityAnalyzer struct{}

func (s *SecurityAnalyzer) Name() string {
	return "security"
}

func (s *SecurityAnalyzer) Analyze(ctx context.Context, snap *model.Snapshot) ([]Finding, error) {
	var findings []Finding

	// Patterns for sensitive files
	sensitivePatterns := []string{
		".env",
		".pem",
		"id_rsa",
		"id_dsa",
		".pypirc",
		".npmrc",
	}

	for _, file := range snap.Files {
		findings = append(findings, s.analyzeRecursive(file, sensitivePatterns)...)
	}

	return findings, nil
}

func (s *SecurityAnalyzer) analyzeRecursive(file *model.FileItem, sensitivePatterns []string) []Finding {
	var findings []Finding

	for _, pattern := range sensitivePatterns {
		if strings.Contains(file.Name, pattern) {
			findings = append(findings, Finding{
				RuleID:     "SEC-001",
				Message:    "Sensitive file detected",
				Path:       file.Path,
				Severity:   SeverityHigh,
				Suggestion: "Remove sensitive files from the repository and use environment variables or a secret manager.",
			})
		}
	}

	for _, child := range file.Children {
		findings = append(findings, s.analyzeRecursive(child, sensitivePatterns)...)
	}

	return findings
}
