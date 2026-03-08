package structure

import (
	"fmt"

	"github.com/repoan/repoan/internal/model"
)

func FindingsFromReport(report *Report, opts Options) []model.Finding {
	if report == nil {
		return nil
	}
	if opts.InstabilityHigh <= 0 {
		opts.InstabilityHigh = 0.85
	}
	if opts.ClusterCouplingHigh <= 0 {
		opts.ClusterCouplingHigh = 0.85
	}
	if opts.ClusterBoundaryLow <= 0 {
		opts.ClusterBoundaryLow = 0.20
	}
	findings := []model.Finding{}

	for _, component := range report.StrongComponents {
		if len(component) <= 1 {
			continue
		}
		sev := model.SevHigh
		cycleSeverity := "high"
		if len(component) >= 4 {
			sev = model.SevCritical
			cycleSeverity = "critical"
		}
		path := component[0]
		finding := model.Finding{
			RuleID:      "REP-STR-001",
			Title:       "Cyclic Module Dependency",
			Message:     fmt.Sprintf("Detected a dependency cycle across %d modules.", len(component)),
			Severity:    sev,
			Category:    model.CatStructure,
			Path:        path,
			Evidence:    fmt.Sprintf("cycle_members=%v", component),
			Remediation: "Break the cycle by introducing directionality boundaries or moving shared abstractions.",
			Tags:        []string{"cycle", "dependency", "architecture"},
			Props: map[string]string{
				"cycle_size":     fmt.Sprintf("%d", len(component)),
				"cycle_severity": cycleSeverity,
			},
		}
		finding.Fingerprint = model.FingerprintStable(finding.RuleID, finding.Path, 0, "module_cycle")
		findings = append(findings, finding)
	}

	for _, m := range report.ModuleMetrics {
		if m.Instability < opts.InstabilityHigh {
			continue
		}
		finding := model.Finding{
			RuleID:      "REP-STR-002",
			Title:       "Highly Unstable Module",
			Message:     fmt.Sprintf("Module '%s' has instability %.2f (fan-in=%d, fan-out=%d).", m.Module, m.Instability, m.FanIn, m.FanOut),
			Severity:    model.SevMedium,
			Category:    model.CatStructure,
			Path:        m.Module,
			Evidence:    fmt.Sprintf("incoming_weight=%d outgoing_weight=%d unresolved_imports=%d", m.IncomingWeight, m.OutgoingWeight, m.UnresolvedImports),
			Remediation: "Reduce outgoing dependencies or increase reuse from other modules to stabilize architecture.",
			Tags:        []string{"instability", "architecture"},
			Props: map[string]string{
				"instability":        fmt.Sprintf("%.3f", m.Instability),
				"fan_in":             fmt.Sprintf("%d", m.FanIn),
				"fan_out":            fmt.Sprintf("%d", m.FanOut),
				"incoming_weight":    fmt.Sprintf("%d", m.IncomingWeight),
				"outgoing_weight":    fmt.Sprintf("%d", m.OutgoingWeight),
				"unresolved_imports": fmt.Sprintf("%d", m.UnresolvedImports),
			},
		}
		finding.Fingerprint = model.FingerprintStable(finding.RuleID, finding.Path, 0, "module_instability")
		findings = append(findings, finding)
	}

	for _, c := range report.ClusterMetrics {
		if c.Modules < 2 {
			continue
		}
		if c.CouplingScore < opts.ClusterCouplingHigh || c.BoundaryStrength > opts.ClusterBoundaryLow {
			continue
		}
		finding := model.Finding{
			RuleID:      "REP-STR-003",
			Title:       "Weak Cluster Boundary",
			Message:     fmt.Sprintf("Cluster '%s' has weak boundaries (coupling=%.2f, boundary=%.2f).", c.Cluster, c.CouplingScore, c.BoundaryStrength),
			Severity:    model.SevLow,
			Category:    model.CatStructure,
			Path:        c.Cluster,
			Evidence:    fmt.Sprintf("modules=%d internal=%d external_in=%d external_out=%d", c.Modules, c.InternalEdges, c.ExternalInEdges, c.ExternalOutEdges),
			Remediation: "Reduce outward coupling and increase cohesion by extracting stable interfaces or moving shared responsibilities.",
			Tags:        []string{"cluster", "boundary", "coupling"},
			Props: map[string]string{
				"cluster":              c.Cluster,
				"modules":              fmt.Sprintf("%d", c.Modules),
				"cohesion_score":       fmt.Sprintf("%.3f", c.CohesionScore),
				"coupling_score":       fmt.Sprintf("%.3f", c.CouplingScore),
				"boundary_strength":    fmt.Sprintf("%.3f", c.BoundaryStrength),
				"directionality_score": fmt.Sprintf("%.3f", c.DirectionalityScore),
			},
		}
		finding.Fingerprint = model.FingerprintStable(finding.RuleID, finding.Path, 0, "weak_cluster_boundary")
		findings = append(findings, finding)
	}

	return findings
}
