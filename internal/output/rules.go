package output

import (
    "encoding/json"
    "fmt"
    "strings"

    "github.com/repoan/repoan/internal/model"
)

// FormatRulesReport formats a `model.RulesReport` into supported formats.
func FormatRulesReport(report *model.RulesReport, format string) (string, error) {
    if report == nil {
        return "", nil
    }
    switch strings.ToLower(format) {
    case "json":
        data, err := json.MarshalIndent(report, "", "  ")
        if err != nil {
            return "", err
        }
        return string(data), nil
    case "text", "md":
        return formatRulesText(report, strings.ToLower(format) == "md"), nil
    default:
        return "", fmt.Errorf("unsupported rules format: %s", format)
    }
}

func formatRulesText(report *model.RulesReport, markdown bool) string {
    var b strings.Builder
    if markdown {
        b.WriteString("# Rules Report\n\n")
    }
    b.WriteString(fmt.Sprintf("Root: %s\n", report.Root))
    b.WriteString(fmt.Sprintf("Generated: %s\n\n", report.GeneratedAt.UTC().Format("2006-01-02T15:04:05Z")))

    b.WriteString("Cluster metrics:\n")
    for _, c := range report.ClusterMetrics {
        b.WriteString(fmt.Sprintf("- %s: modules=%d cohesion=%.3f coupling=%.3f boundary=%.3f directionality=%.3f instability=%.3f in-cycle=%d\n",
            c.Cluster, c.Modules, c.CohesionScore, c.CouplingScore, c.BoundaryStrength, c.DirectionalityScore, c.Instability, c.InCycleModules))
    }

    if len(report.StrongComponents) > 0 {
        b.WriteString("\nStrong components:\n")
        for _, comp := range report.StrongComponents {
            b.WriteString("- ")
            b.WriteString(strings.Join(comp, ", "))
            b.WriteString("\n")
        }
    }

    return b.String()
}
