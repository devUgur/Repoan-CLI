package output

import (
	"encoding/json"
	"github.com/repoan/repoan/internal/analyze"
)

// SarifReport is a basic SARIF structure.
type SarifReport struct {
	Version string     `json:"version"`
	Runs    []SarifRun `json:"runs"`
}

type SarifRun struct {
	Tool    SarifTool    `json:"tool"`
	Results []SarifResult `json:"results"`
}

type SarifTool struct {
	Driver SarifDriver `json:"driver"`
}

type SarifDriver struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SarifResult struct {
	RuleID  string       `json:"ruleId"`
	Message SarifMessage `json:"message"`
	Level   string       `json:"level"`
	Locations []SarifLocation `json:"locations"`
}

type SarifMessage struct {
	Text string `json:"text"`
}

type SarifLocation struct {
	PhysicalLocation SarifPhysicalLocation `json:"physicalLocation"`
}

type SarifPhysicalLocation struct {
	ArtifactLocation SarifArtifactLocation `json:"artifactLocation"`
}

type SarifArtifactLocation struct {
	Uri string `json:"uri"`
}

func FormatSarif(findings []analyze.Finding, toolName, toolVersion string) (string, error) {
	report := SarifReport{
		Version: "2.1.0",
		Runs: []SarifRun{
			{
				Tool: SarifTool{
					Driver: SarifDriver{
						Name:    toolName,
						Version: toolVersion,
					},
				},
				Results: make([]SarifResult, 0, len(findings)),
			},
		},
	}

	for _, f := range findings {
		level := "warning"
		if f.Severity == analyze.SeverityHigh || f.Severity == analyze.SeverityCritical {
			level = "error"
		} else if f.Severity == analyze.SeverityInfo {
			level = "note"
		}

		result := SarifResult{
			RuleID: f.RuleID,
			Message: SarifMessage{
				Text: f.Message,
			},
			Level: level,
			Locations: []SarifLocation{
				{
					PhysicalLocation: SarifPhysicalLocation{
						ArtifactLocation: SarifArtifactLocation{
							Uri: f.Path,
						},
					},
				},
			},
		}
		report.Runs[0].Results = append(report.Runs[0].Results, result)
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
