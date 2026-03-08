package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the tool's main configuration.
type Config struct {
	Version  int            `yaml:"version"`
	Scan     ScanConfig     `yaml:"scan"`
	Output   OutputConfig   `yaml:"output"`
	Analysis AnalysisConfig `yaml:"analysis"`
}

type ScanConfig struct {
	RespectGitignore bool     `yaml:"respect_gitignore"`
	MaxDepth         int      `yaml:"max_depth"`
	Ignore           []string `yaml:"ignore"`
}

type OutputConfig struct {
	Dir           string `yaml:"dir"`
	DefaultFormat string `yaml:"default_format"`
}

type AnalysisConfig struct {
	EnabledRules []string        `yaml:"enabled_rules"`
	FailOn       string          `yaml:"fail_on"`
	Structure    StructureConfig `yaml:"structure"`
}

type StructureConfig struct {
	TopTokens           int     `yaml:"top_tokens"`
	IncludeTests        bool    `yaml:"include_tests"`
	InstabilityHigh     float64 `yaml:"instability_high"`
	ClusterCouplingHigh float64 `yaml:"cluster_coupling_high"`
	ClusterBoundaryLow  float64 `yaml:"cluster_boundary_low"`
}

// BuiltInIgnorePatterns returns the default set of ignore patterns used by scan-related commands.
func BuiltInIgnorePatterns() []string {
	return []string{
		".git/",
		"node_modules/",
		"dist/",
		"build/",
		"target/",
		"vendor/",
		"bin/",
		"obj/",
		"out/",
		"coverage/",
		"tmp/",
		"temp/",
		".venv/",
		"venv/",
		"__pycache__/",
		".next/",
		".nuxt/",
		".output/",
		".svelte-kit/",
		".angular/",
		".cache/",
		".parcel-cache/",
		".turbo/",
		".pnpm-store/",
		".yarn/",
		".terraform/",
		".repoan/reports/",
		".repoan/cache/",
		".repoan/logs/",
		".DS_Store",
		"*.exe",
		"*.dll",
	}
}

// EffectiveIgnorePatterns merges built-in and user-provided ignores with stable ordering.
func EffectiveIgnorePatterns(userPatterns []string) []string {
	merged := make([]string, 0, len(BuiltInIgnorePatterns())+len(userPatterns))
	seen := make(map[string]struct{})

	appendUnique := func(patterns []string) {
		for _, p := range patterns {
			if p == "" {
				continue
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			merged = append(merged, p)
		}
	}

	appendUnique(BuiltInIgnorePatterns())
	appendUnique(userPatterns)

	return merged
}

// Default returns a new Config with default values.
func Default() *Config {
	return &Config{
		Version: 1,
		Scan: ScanConfig{
			RespectGitignore: true,
			MaxDepth:         0,
			Ignore:           BuiltInIgnorePatterns(),
		},
		Output: OutputConfig{
			Dir:           ".repoan/reports",
			DefaultFormat: "md",
		},
		Analysis: AnalysisConfig{
			EnabledRules: []string{
				"large-files",
				"security",
				"structure",
			},
			FailOn: "high",
			Structure: StructureConfig{
				TopTokens:           25,
				IncludeTests:        false,
				InstabilityHigh:     0.85,
				ClusterCouplingHigh: 0.85,
				ClusterBoundaryLow:  0.20,
			},
		},
	}
}

// Load loads the configuration from a file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves the configuration to a file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
