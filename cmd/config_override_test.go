package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStructureAndAnalyzeRespectConfigThresholdOverrides(t *testing.T) {
	repoRoot := findRepoRoot(t)
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, "repoan.config.yml")
	structureOut := filepath.Join(tmp, "structure.sarif")
	analyzeOut := filepath.Join(tmp, "analyze.json")

	configYAML := `version: 1
scan:
  respect_gitignore: true
  max_depth: 0
  ignore: []
analysis:
  enabled_rules:
    - "structure"
  fail_on: "none"
  structure:
    top_tokens: 20
    include_tests: false
    instability_high: 0.99
    cluster_coupling_high: 0.10
    cluster_boundary_low: 1.00
output:
  dir: ".repoan/reports"
  default_format: "json"
`

	if err := os.WriteFile(cfg, []byte(configYAML), 0644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	cmdStructure := exec.Command("go", "run", ".", "--config", cfg, "structure", "--here", "--format", "sarif", "--out", structureOut)
	cmdStructure.Dir = repoRoot
	if out, err := cmdStructure.CombinedOutput(); err != nil {
		t.Fatalf("structure with config failed: %v\noutput:\n%s", err, string(out))
	}

	dataStructure, err := os.ReadFile(structureOut)
	if err != nil {
		t.Fatalf("read structure out failed: %v", err)
	}
	if !strings.Contains(string(dataStructure), "REP-STR-003") {
		t.Fatalf("expected structure sarif to include REP-STR-003 with permissive cluster thresholds")
	}

	cmdAnalyze := exec.Command("go", "run", ".", "--config", cfg, "analyze", "--format", "json", "--out", analyzeOut)
	cmdAnalyze.Dir = repoRoot
	if out, err := cmdAnalyze.CombinedOutput(); err != nil {
		t.Fatalf("analyze with config failed: %v\noutput:\n%s", err, string(out))
	}

	dataAnalyze, err := os.ReadFile(analyzeOut)
	if err != nil {
		t.Fatalf("read analyze out failed: %v", err)
	}
	if !strings.Contains(string(dataAnalyze), "REP-STR-003") {
		t.Fatalf("expected analyze json to include REP-STR-003 with permissive cluster thresholds")
	}
}
