package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the tool's main configuration.
type Config struct {
	Version int `yaml:"version"`
	Scan    ScanConfig `yaml:"scan"`
	Output  OutputConfig `yaml:"output"`
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
	EnabledRules []string `yaml:"enabled_rules"`
	FailOn       string   `yaml:"fail_on"`
}

// Default returns a new Config with default values.
func Default() *Config {
	return &Config{
		Version: 1,
		Scan: ScanConfig{
			RespectGitignore: true,
			MaxDepth:         0,
			Ignore: []string{
				".git/",
				"node_modules/",
				"dist/",
				"build/",
				"target/",
				".venv/",
				"vendor/",
			},
		},
		Output: OutputConfig{
			Dir:           ".repoan/reports",
			DefaultFormat: "md",
		},
		Analysis: AnalysisConfig{
			EnabledRules: []string{
				"large-files",
				"security",
			},
			FailOn: "high",
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
