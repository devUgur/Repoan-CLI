package config

import "testing"

func TestEffectiveIgnorePatterns_MergesAndDeduplicates(t *testing.T) {
	merged := EffectiveIgnorePatterns([]string{"custom/", "node_modules/", "custom/"})

	seenCustom := 0
	seenNodeModules := 0
	for _, p := range merged {
		if p == "custom/" {
			seenCustom++
		}
		if p == "node_modules/" {
			seenNodeModules++
		}
	}

	if seenCustom != 1 {
		t.Fatalf("expected custom/ once, got %d", seenCustom)
	}
	if seenNodeModules != 1 {
		t.Fatalf("expected node_modules/ once, got %d", seenNodeModules)
	}

	if len(merged) < len(BuiltInIgnorePatterns())+1 {
		t.Fatalf("expected merged to include built-ins plus custom entries, got len=%d", len(merged))
	}
}

func TestDefaultStructureThresholds(t *testing.T) {
	cfg := Default()
	if cfg.Analysis.Structure.ClusterCouplingHigh <= 0 || cfg.Analysis.Structure.ClusterCouplingHigh > 1 {
		t.Fatalf("expected cluster coupling threshold in (0,1], got %f", cfg.Analysis.Structure.ClusterCouplingHigh)
	}
	if cfg.Analysis.Structure.ClusterBoundaryLow <= 0 {
		t.Fatalf("expected positive cluster boundary low threshold, got %f", cfg.Analysis.Structure.ClusterBoundaryLow)
	}
}
