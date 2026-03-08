package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStructureCommand_JSONOut(t *testing.T) {
	repoRoot := findRepoRoot(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "structure.json")

	cmd := exec.Command("go", "run", ".", "structure", "--here", "--format", "json", "--out", out, "--top-tokens", "5", "--include-tests")
	cmd.Dir = repoRoot
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("structure command failed: %v\noutput:\n%s", err, string(b))
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}
	content := string(data)
	for _, mustContain := range []string{"\"summary\"", "\"cluster_metrics\"", "\"layer_candidates\"", "\"resolved_imports\""} {
		if !strings.Contains(content, mustContain) {
			t.Fatalf("expected output to contain %s", mustContain)
		}
	}
}

func TestStructureCommand_MermaidOut(t *testing.T) {
	repoRoot := findRepoRoot(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "deps.mmd")

	cmd := exec.Command("go", "run", ".", "structure", "--here", "--format", "mermaid", "--out", out)
	cmd.Dir = repoRoot
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("structure mermaid command failed: %v\noutput:\n%s", err, string(b))
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("could not read mermaid file: %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "graph TD") {
		t.Fatalf("expected mermaid graph TD header")
	}
}

func TestStructureCommand_MermaidClusterOut(t *testing.T) {
	repoRoot := findRepoRoot(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "deps-cluster.mmd")

	cmd := exec.Command("go", "run", ".", "structure", "--here", "--format", "mermaid-cluster", "--out", out)
	cmd.Dir = repoRoot
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("structure cluster mermaid command failed: %v\noutput:\n%s", err, string(b))
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("could not read cluster mermaid file: %v", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "graph TD") {
		t.Fatalf("expected mermaid graph TD header")
	}
}

func TestStructureCommand_RawDependenciesToggle(t *testing.T) {
	repoRoot := findRepoRoot(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "structure-raw.json")

	cmd := exec.Command("go", "run", ".", "structure", "--here", "--format", "json", "--out", out, "--raw-dependencies", "--raw-dependencies-limit", "5")
	cmd.Dir = repoRoot
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("structure raw dependencies command failed: %v\noutput:\n%s", err, string(b))
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "\"raw_dependencies\"") {
		t.Fatalf("expected output to include raw_dependencies field when enabled")
	}
}

func TestStructureCommand_RawDependenciesDisabledByDefault(t *testing.T) {
	repoRoot := findRepoRoot(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "structure-default.json")

	cmd := exec.Command("go", "run", ".", "structure", "--here", "--format", "json", "--out", out)
	cmd.Dir = repoRoot
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("structure default command failed: %v\noutput:\n%s", err, string(b))
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "\"raw_dependencies\"") {
		t.Fatalf("did not expect raw_dependencies field by default")
	}
}

func TestStructureCommand_RawDependenciesOnlyUnresolved(t *testing.T) {
	repoRoot := findRepoRoot(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "structure-unresolved.json")

	cmd := exec.Command("go", "run", ".", "structure", "--here", "--format", "json", "--out", out, "--raw-dependencies", "--raw-dependencies-only-unresolved", "--raw-dependencies-limit", "20")
	cmd.Dir = repoRoot
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("structure unresolved-only command failed: %v\noutput:\n%s", err, string(b))
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}

	var parsed struct {
		RawDependencies []struct {
			Resolved bool `json:"resolved"`
		} `json:"raw_dependencies"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("could not parse json output: %v", err)
	}

	for _, dep := range parsed.RawDependencies {
		if dep.Resolved {
			t.Fatalf("expected unresolved-only raw dependencies, found resolved entry")
		}
	}
}

func TestStructureCommand_RawDependenciesOnlyExternal(t *testing.T) {
	repoRoot := findRepoRoot(t)
	tmp := t.TempDir()
	out := filepath.Join(tmp, "structure-external.json")

	cmd := exec.Command("go", "run", ".", "structure", "--here", "--format", "json", "--out", out, "--raw-dependencies", "--raw-dependencies-only-external", "--raw-dependencies-limit", "50")
	cmd.Dir = repoRoot
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("structure external-only command failed: %v\noutput:\n%s", err, string(b))
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("could not read output file: %v", err)
	}

	var parsed struct {
		RawDependencies []struct {
			External bool `json:"external"`
		} `json:"raw_dependencies"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("could not parse json output: %v", err)
	}

	for _, dep := range parsed.RawDependencies {
		if !dep.External {
			t.Fatalf("expected external-only raw dependencies, found internal entry")
		}
	}
}

func TestStructureCommand_RawDependenciesMinConfidenceValidation(t *testing.T) {
	repoRoot := findRepoRoot(t)

	cmd := exec.Command("go", "run", ".", "structure", "--here", "--raw-dependencies", "--raw-dependencies-min-confidence", "1.5")
	cmd.Dir = repoRoot
	b, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected command to fail for invalid min-confidence")
	}
	if !strings.Contains(string(b), "--raw-dependencies-min-confidence") {
		t.Fatalf("expected validation error to mention raw-dependencies-min-confidence, got output: %s", string(b))
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	root := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("could not detect repository root from %s", wd)
	}
	return root
}
