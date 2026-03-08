package structure

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/repoan/repoan/internal/model"
	"github.com/repoan/repoan/internal/scan"
)

func TestAnalyze_BuildsModuleMetricsAndCycles(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "go.mod"), "module github.com/acme/demo\n\ngo 1.25.0\n")
	mustWriteFile(t, filepath.Join(tmp, "cmd", "main.go"), "package main\nimport _ \"github.com/acme/demo/internal/a\"\nfunc main(){}\n")
	mustWriteFile(t, filepath.Join(tmp, "internal", "a", "a.go"), "package a\nimport _ \"github.com/acme/demo/internal/b\"\n")
	mustWriteFile(t, filepath.Join(tmp, "internal", "b", "b.go"), "package b\nimport _ \"github.com/acme/demo/internal/a\"\n")

	tree, err := scan.Scan(scan.ScanOptions{Root: tmp})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	snap := &model.Snapshot{
		RepoRoot:  tmp,
		Timestamp: time.Now(),
		Files:     []*model.FileItem{tree},
	}

	report, err := Analyze(tmp, snap, Options{TopTokens: 10, InstabilityHigh: 0.8})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}

	if report.Summary.Modules == 0 {
		t.Fatalf("expected modules in report")
	}
	if report.Summary.Cycles == 0 {
		t.Fatalf("expected at least one cycle")
	}
	if len(report.ModuleEdges) == 0 {
		t.Fatalf("expected dependency edges")
	}
	if len(report.ClusterMetrics) == 0 {
		t.Fatalf("expected cluster metrics")
	}

	foundInternal := false
	for _, m := range report.ModuleMetrics {
		if m.Module == "internal/a" || m.Module == "internal/b" {
			foundInternal = true
			if !m.InCycle {
				t.Fatalf("expected %s module to be marked in cycle", m.Module)
			}
		}
	}
	if !foundInternal {
		t.Fatalf("expected internal sub-module metrics")
	}

	foundInternalCluster := false
	for _, c := range report.ClusterMetrics {
		if c.Cluster == "internal" {
			foundInternalCluster = true
			if c.Modules < 2 {
				t.Fatalf("expected internal cluster to include at least 2 modules")
			}
		}
	}
	if !foundInternalCluster {
		t.Fatalf("expected internal cluster metrics")
	}
}

func TestFindingsFromReport_ProducesStructureFindings(t *testing.T) {
	report := &Report{
		StrongComponents: [][]string{{"a", "b"}},
		ModuleMetrics: []ModuleMetrics{
			{Module: "api", FanIn: 0, FanOut: 5, Instability: 0.91},
		},
		ClusterMetrics: []ClusterMetrics{
			{Cluster: "api", Modules: 3, CouplingScore: 0.92, BoundaryStrength: 0.10, CohesionScore: 0.15, InternalEdges: 1, ExternalInEdges: 3, ExternalOutEdges: 6},
		},
	}

	findings := FindingsFromReport(report, Options{InstabilityHigh: 0.85, ClusterCouplingHigh: 0.85, ClusterBoundaryLow: 0.20})
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
}

func TestAnalyze_ResolvesTSConfigAlias(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "tsconfig.json"), `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@app/*": ["src/*"]
    }
  }
}`)
	mustWriteFile(t, filepath.Join(tmp, "src", "lib", "helper.ts"), "export const x = 1;\n")
	mustWriteFile(t, filepath.Join(tmp, "src", "feature", "service.ts"), "import { x } from \"@app/lib/helper\";\nexport const y = x;\n")

	tree, err := scan.Scan(scan.ScanOptions{Root: tmp})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	snap := &model.Snapshot{
		RepoRoot:  tmp,
		Timestamp: time.Now(),
		Files:     []*model.FileItem{tree},
	}

	report, err := Analyze(tmp, snap, Options{TopTokens: 10, IncludeTests: true, InstabilityHigh: 0.8})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}

	if report.Summary.ResolvedImports == 0 {
		t.Fatalf("expected resolved imports > 0")
	}

	foundAliasEdge := false
	for _, edge := range report.ModuleEdges {
		if edge.Source == "js-regex" && edge.Resolved {
			foundAliasEdge = true
			break
		}
	}
	if !foundAliasEdge {
		t.Fatalf("expected at least one resolved TS alias edge")
	}

	if len(report.LayerCandidates) == 0 || len(report.LayerCandidates[0].Reasons) == 0 {
		t.Fatalf("expected explainable layer candidate reasons")
	}
}

func TestAnalyze_RawDependenciesOptionalAndLimited(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "pkg", "b.go"), "package pkg\n")
	mustWriteFile(t, filepath.Join(tmp, "pkg", "a.go"), "package pkg\nimport _ \"./b\"\n")

	tree, err := scan.Scan(scan.ScanOptions{Root: tmp})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	snap := &model.Snapshot{RepoRoot: tmp, Timestamp: time.Now(), Files: []*model.FileItem{tree}}

	withoutRaw, err := Analyze(tmp, snap, Options{TopTokens: 5})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if len(withoutRaw.RawDependencies) != 0 {
		t.Fatalf("expected raw dependencies to be disabled by default")
	}

	withRaw, err := Analyze(tmp, snap, Options{TopTokens: 5, RawDependencies: RawDependencyOptions{Enabled: true, Limit: 1}})
	if err != nil {
		t.Fatalf("analyze with raw dependencies failed: %v", err)
	}
	if len(withRaw.RawDependencies) != 1 {
		t.Fatalf("expected raw dependencies to be limited to 1, got %d", len(withRaw.RawDependencies))
	}
	if withRaw.RawDependencies[0].FromFile == "" || withRaw.RawDependencies[0].SourceImport == "" {
		t.Fatalf("expected populated raw dependency fields")
	}
	if withRaw.RawDepSummary == nil {
		t.Fatalf("expected raw dependency summary when raw deps are enabled")
	}
}

func TestAnalyze_RawDependenciesOnlyUnresolvedFilter(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "src", "dep.ts"), "export const dep = 1;\n")
	mustWriteFile(t, filepath.Join(tmp, "src", "main.ts"), "import { dep } from \"./dep\";\nimport { missing } from \"./missing\";\nexport const x = dep;\n")

	tree, err := scan.Scan(scan.ScanOptions{Root: tmp})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	snap := &model.Snapshot{RepoRoot: tmp, Timestamp: time.Now(), Files: []*model.FileItem{tree}}

	report, err := Analyze(tmp, snap, Options{TopTokens: 5, RawDependencies: RawDependencyOptions{Enabled: true, OnlyUnresolved: true}})
	if err != nil {
		t.Fatalf("analyze with unresolved-only raw dependencies failed: %v", err)
	}

	if len(report.RawDependencies) == 0 {
		t.Fatalf("expected unresolved raw dependencies to be present")
	}
	for _, dep := range report.RawDependencies {
		if dep.Resolved {
			t.Fatalf("expected unresolved-only filter to exclude resolved dependencies")
		}
	}
}

func TestAnalyze_RawDependenciesOnlyExternalAndComposableFilters(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "src", "dep.ts"), "export const dep = 1;\n")
	mustWriteFile(t, filepath.Join(tmp, "src", "main.ts"), "import { dep } from \"./dep\";\nimport { missing } from \"./missing\";\nimport React from \"react\";\nexport const x = dep + (React ? 1 : 0);\n")

	tree, err := scan.Scan(scan.ScanOptions{Root: tmp})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	snap := &model.Snapshot{RepoRoot: tmp, Timestamp: time.Now(), Files: []*model.FileItem{tree}}

	report, err := Analyze(tmp, snap, Options{TopTokens: 5, RawDependencies: RawDependencyOptions{Enabled: true, OnlyExternal: true, OnlyUnresolved: true, MinConfidence: 0.5}})
	if err != nil {
		t.Fatalf("analyze with composable raw dependency filters failed: %v", err)
	}

	if len(report.RawDependencies) == 0 {
		t.Fatalf("expected filtered raw dependencies to be present")
	}
	for _, dep := range report.RawDependencies {
		if dep.Resolved {
			t.Fatalf("expected unresolved-only filter to exclude resolved dependencies")
		}
		if !dep.External {
			t.Fatalf("expected external-only filter to exclude internal dependencies")
		}
		if dep.Confidence < 0.5 {
			t.Fatalf("expected min-confidence filter to exclude low-confidence dependencies")
		}
	}
}

func TestAnalyze_RawDependenciesMinConfidenceValidation(t *testing.T) {
	tmp := t.TempDir()

	mustWriteFile(t, filepath.Join(tmp, "main.go"), "package main\nfunc main() {}\n")
	tree, err := scan.Scan(scan.ScanOptions{Root: tmp})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	snap := &model.Snapshot{RepoRoot: tmp, Timestamp: time.Now(), Files: []*model.FileItem{tree}}

	if _, err := Analyze(tmp, snap, Options{TopTokens: 5, RawDependencies: RawDependencyOptions{Enabled: true, MinConfidence: -0.1}}); err == nil {
		t.Fatalf("expected validation error for negative min confidence")
	}
	if _, err := Analyze(tmp, snap, Options{TopTokens: 5, RawDependencies: RawDependencyOptions{Enabled: true, MinConfidence: 1.1}}); err == nil {
		t.Fatalf("expected validation error for min confidence above 1")
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
