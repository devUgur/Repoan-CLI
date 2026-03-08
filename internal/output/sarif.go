package output

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/repoan/repoan/internal/model"
)

type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results,omitempty"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri,omitempty"`
	Version        string      `json:"version,omitempty"`
	Rules          []sarifRule `json:"rules,omitempty"`
}

type sarifRule struct {
	ID               string        `json:"id"`
	Name             string        `json:"name,omitempty"`
	ShortDescription *sarifMessage `json:"shortDescription,omitempty"`
	Help             *sarifMessage `json:"help,omitempty"`
}

type sarifResult struct {
	RuleID     string          `json:"ruleId"`
	Level      string          `json:"level,omitempty"`
	Message    sarifMessage    `json:"message"`
	Locations  []sarifLocation `json:"locations,omitempty"`
	Properties map[string]any  `json:"properties,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion          `json:"region,omitempty"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

func severityToSarifLevel(s model.Severity) string {
	switch s {
	case model.SevHigh, model.SevCritical:
		return "error"
	case model.SevMedium:
		return "warning"
	default:
		return "note"
	}
}

// FormatSarif generates a SARIF 2.1.0 string from findings.
func FormatSarif(findings []model.Finding, toolName, toolVersion string) (string, error) {
	ruleMap := map[string]sarifRule{}
	results := make([]sarifResult, 0, len(findings))

	for _, f := range findings {
		if _, ok := ruleMap[f.RuleID]; !ok {
			ruleMap[f.RuleID] = sarifRule{
				ID:               f.RuleID,
				Name:             f.Title,
				ShortDescription: &sarifMessage{Text: f.Title},
				Help:             &sarifMessage{Text: f.Remediation},
			}
		}

		loc := sarifLocation{
			PhysicalLocation: sarifPhysicalLocation{
				ArtifactLocation: sarifArtifactLocation{URI: model.NormalizePath(f.Path)},
			},
		}
		if f.Line > 0 || f.Column > 0 {
			loc.PhysicalLocation.Region = &sarifRegion{StartLine: f.Line, StartColumn: f.Column}
		}

		props := map[string]any{
			"category":    string(f.Category),
			"severity":    string(f.Severity),
			"fingerprint": f.Fingerprint,
			"generatedAt": time.Now().UTC().Format(time.RFC3339),
		}
		if f.Evidence != "" {
			props["evidence"] = f.Evidence
		}
		if len(f.Tags) > 0 {
			props["tags"] = f.Tags
		}
		for k, v := range f.Props {
			props[k] = v
		}

		results = append(results, sarifResult{
			RuleID: f.RuleID,
			Level:  severityToSarifLevel(f.Severity),
			Message: sarifMessage{
				Text: f.Message,
			},
			Locations:  []sarifLocation{loc},
			Properties: props,
		})
	}

	rules := make([]sarifRule, 0, len(ruleMap))
	for _, r := range ruleMap {
		rules = append(rules, r)
	}
	// Sort rules by ID for determinism
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ID < rules[j].ID
	})

	log := sarifLog{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           toolName,
						Version:        toolVersion,
						InformationURI: "https://github.com/repoan/repoan",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}

	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
