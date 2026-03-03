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

func (s *SecurityAnalyzer) Analyze(ctx context.Context, snap *model.Snapshot) ([]model.Finding, error) {
	var findings []model.Finding

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

func (s *SecurityAnalyzer) analyzeRecursive(file *model.FileItem, sensitivePatterns []string) []model.Finding {
	var findings []model.Finding

	for _, pattern := range sensitivePatterns {
		if strings.Contains(file.Name, pattern) {
			finding := model.Finding{
				RuleID:      "REP-SEC-001",
				Title:       "Sensitive File Detected",
				Message:     "A file that may contain credentials or sensitive information was found.",
				Path:        file.Path,
				Severity:    model.SevHigh,
				Category:    model.CatSecurity,
				Remediation: "Remove sensitive files from the repository and use environment variables or a secret manager.",
			}
			finding.Fingerprint = model.FingerprintStable(finding.RuleID, finding.Path, 0, pattern)
			findings = append(findings, finding)
		}
	}

	for _, child := range file.Children {
		findings = append(findings, s.analyzeRecursive(child, sensitivePatterns)...)
	}

	return findings
}
