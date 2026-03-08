package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/repoan/repoan/internal/structure"
)

func FormatStructureReport(report *structure.Report, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	case "md", "text":
		return formatStructureText(report, strings.ToLower(format) == "md"), nil
	case "mermaid":
		return formatStructureMermaid(report), nil
	case "mermaid-cluster":
		return formatStructureClusterMermaid(report), nil
	default:
		return "", fmt.Errorf("unsupported structure format: %s", format)
	}
}

func formatStructureText(report *structure.Report, markdown bool) string {
	if report == nil {
		return ""
	}
	var b strings.Builder
	if markdown {
		b.WriteString("# Repo Structure Report\n\n")
	}
	b.WriteString(fmt.Sprintf("Files: %d\n", report.Summary.Files))
	b.WriteString(fmt.Sprintf("Modules: %d\n", report.Summary.Modules))
	b.WriteString(fmt.Sprintf("Dependency edges: %d\n", report.Summary.DependencyEdges))
	b.WriteString(fmt.Sprintf("Resolved imports: %d\n", report.Summary.ResolvedImports))
	b.WriteString(fmt.Sprintf("Unresolved imports: %d\n", report.Summary.UnresolvedImports))
	b.WriteString(fmt.Sprintf("Cycles: %d\n\n", report.Summary.Cycles))

	b.WriteString("Top modules by instability:\n")
	for i, m := range report.ModuleMetrics {
		if i >= 10 {
			break
		}
		b.WriteString(fmt.Sprintf("- %s: instability=%.3f fan-in=%d fan-out=%d in-w=%d out-w=%d unresolved=%d cycle=%t\n", m.Module, m.Instability, m.FanIn, m.FanOut, m.IncomingWeight, m.OutgoingWeight, m.UnresolvedImports, m.InCycle))
	}

	b.WriteString("\nTop structural tokens:\n")
	for i, t := range report.Tokens {
		if i >= 10 {
			break
		}
		b.WriteString(fmt.Sprintf("- %s (%s): weight=%.3f count=%d spread=%.3f\n", t.Name, t.Kind, t.Weight, t.Count, t.Spread))
	}

	b.WriteString("\nLayer candidates:\n")
	for i, c := range report.LayerCandidates {
		if i >= 10 {
			break
		}
		b.WriteString(fmt.Sprintf("- %s -> %s (confidence=%.3f)\n", c.Module, c.Role, c.Confidence))
		if len(c.Reasons) > 0 {
			b.WriteString(fmt.Sprintf("  reasons: %s\n", strings.Join(c.Reasons, "; ")))
		}
	}

	b.WriteString("\nCluster metrics:\n")
	for i, c := range report.ClusterMetrics {
		if i >= 10 {
			break
		}
		b.WriteString(fmt.Sprintf("- %s: modules=%d cohesion=%.3f coupling=%.3f boundary=%.3f directionality=%.3f instability=%.3f in-cycle=%d\n",
			c.Cluster, c.Modules, c.CohesionScore, c.CouplingScore, c.BoundaryStrength, c.DirectionalityScore, c.Instability, c.InCycleModules))
	}

	return b.String()
}

func formatStructureMermaid(report *structure.Report) string {
	if report == nil {
		return "graph TD\n"
	}
	var b strings.Builder
	b.WriteString("graph TD\n")
	for _, edge := range report.ModuleEdges {
		b.WriteString(fmt.Sprintf("  %s -->|%d| %s\n", sanitizeMermaidID(edge.From), edge.Weight, sanitizeMermaidID(edge.To)))
	}
	return b.String()
}

func formatStructureClusterMermaid(report *structure.Report) string {
	if report == nil {
		return "graph TD\n"
	}
	clusterEdges := map[string]map[string]int{}
	for _, edge := range report.ModuleEdges {
		from := clusterName(edge.From)
		to := clusterName(edge.To)
		if from == to {
			continue
		}
		if _, ok := clusterEdges[from]; !ok {
			clusterEdges[from] = map[string]int{}
		}
		clusterEdges[from][to] += edge.Weight
	}

	var b strings.Builder
	b.WriteString("graph TD\n")
	froms := make([]string, 0, len(clusterEdges))
	for from := range clusterEdges {
		froms = append(froms, from)
	}
	sort.Strings(froms)
	for _, from := range froms {
		tos := make([]string, 0, len(clusterEdges[from]))
		for to := range clusterEdges[from] {
			tos = append(tos, to)
		}
		sort.Strings(tos)
		for _, to := range tos {
			w := clusterEdges[from][to]
			b.WriteString(fmt.Sprintf("  %s -->|%d| %s\n", sanitizeMermaidID(from), w, sanitizeMermaidID(to)))
		}
	}
	return b.String()
}

func clusterName(module string) string {
	module = strings.TrimSpace(module)
	if module == "" {
		return "root"
	}
	parts := strings.Split(module, "/")
	if len(parts) == 0 || parts[0] == "" {
		return "root"
	}
	return parts[0]
}

func sanitizeMermaidID(id string) string {
	replacer := strings.NewReplacer("-", "_", "/", "_", ".", "_", " ", "_")
	id = replacer.Replace(id)
	if id == "" {
		return "root"
	}
	return id
}
