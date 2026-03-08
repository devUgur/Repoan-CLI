# Repoan-CLI 🚀

[![Go Report Card](https://goreportcard.com/badge/github.com/repoan/repoan)](https://goreportcard.com/report/github.com/repoan/repoan)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/repoan/repoan)](https://github.com/repoan/repoan/releases)

**Repoan** is a professional, fast, and cross-platform CLI tool for directory structure snapshots and repository analysis. It helps developers and teams document codebases, prepare repository contexts for LLMs, and integrate security and hygiene checks into CI/CD pipelines.

> 💡 **Repoan is lightweight, language-agnostic, and offline-first.** It serves as an efficient pre-scanner for local development and CI/CD gates, without vendor lock-in or SaaS requirements.

---

## ✨ Features

- **Repository-Aware Scanning:** Automatically detects the Git root.
- **Intelligent Filtering:** Respects `.gitignore` and custom ignore patterns from `.repoan/config.yml`.
- **Flexible Formats:** Output directory structures as Text, Markdown, or JSON.
- **Repository Snapshots:** Create detailed JSON snapshots including file sizes and metadata.
- **Integrated Analysis:**
    - **Security:** Detects sensitive files (e.g., `.env`, `.pem`, private keys).
    - **Hygiene:** Identifies oversized files or binaries.
- **CI/CD Ready:** Exports results in **SARIF format** (natively supported by GitHub Code Scanning) and provides `--fail-on` thresholds.
- **Centralized Configuration:** Everything neatly organized in a `.repoan/` directory.
- **Deterministic:** Identical repositories produce identical results – perfect for stable CI.

---

## 🚀 Installation

### 1. Via Go (for developers)
```bash
go install github.com/repoan/repoan@latest
```
Ensure that your `GOPATH/bin` is in your system PATH.

### 2. Binaries (Windows, macOS, Linux)
Download the appropriate binary for your system from the [Releases page](https://github.com/repoan/repoan/releases) and add it to your PATH.

---

## 🛠️ Getting Started

### Initialization
Prepare your repository for Repoan. This creates a `.repoan/` folder with default settings.
```bash
repoan init
```

### Show Directory Structure (Tree)
```bash
# Standard tree view (respects .gitignore)
repoan tree

# Clean structure for docs (directories only)
repoan tree --dirs-only

# As Markdown for documentation
repoan tree --format md

# Limited depth
repoan tree --max-depth 2

# Redirect to file (ASCII is auto-enabled when redirected)
repoan tree --dirs-only > structure.txt
```

### Create Repository Snapshot
```bash
repoan snapshot --out repo-snapshot.json
```

### Run Analysis
```bash
# Quick check (text output)
repoan analyze

# Export for GitHub Code Scanning (SARIF)
repoan analyze --format sarif --out results.sarif
```

### Analyze Repository Structure
```bash
# Human-readable architecture summary
repoan structure

# Machine-readable report
repoan structure --format json --out structure-report.json

# JSON report with raw file-level dependency traces (debug)
repoan structure --format json --raw-dependencies --raw-dependencies-limit 500 --out structure-debug.json

# Show only unresolved dependency traces for resolver triage
repoan structure --format json --raw-dependencies --raw-dependencies-only-unresolved --out unresolved-debug.json

# Focus on uncertain external dependencies only
repoan structure --format json --raw-dependencies --raw-dependencies-only-external --raw-dependencies-min-confidence 0.4 --out external-low-confidence-debug.json

# Dependency graph for docs
repoan structure --format mermaid --out deps.mmd

# Cluster-level dependency graph
repoan structure --format mermaid-cluster --out deps-cluster.mmd
```

`--raw-dependencies` adds a `raw_dependencies` section in JSON output so you can inspect resolver decisions per import. Use `--raw-dependencies-only-unresolved`, `--raw-dependencies-only-external`, and `--raw-dependencies-min-confidence` to compose targeted debug views.

---

## ⚙️ Configuration (`.repoan/config.yml`)

Repoan is controlled via a central YAML file. Example:

```yaml
version: 1
scan:
  respect_gitignore: true
  max_depth: 0
  ignore:
    - ".git/"
    - "node_modules/"
    - "dist/"
analysis:
  enabled_rules:
    - "security"
    - "hygiene"
    - "structure"
  fail_on: "high"
  structure:
    top_tokens: 25
    include_tests: false
    instability_high: 0.85
    cluster_coupling_high: 0.85
    cluster_boundary_low: 0.20
output:
  dir: ".repoan/reports"
  default_format: "md"
```

---

## 🤖 CI/CD Integration

Repoan is optimized for pipeline usage. Example for GitHub Actions:

```yaml
- name: Run Repoan Analysis
  run: repoan analyze --format sarif --out repoan-results.sarif --fail-on high

- name: Upload SARIF results
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: repoan-results.sarif
```

---

## 🤝 Contributing

Contributions are welcome! Check our [CONTRIBUTING.md](CONTRIBUTING.md) for details on the development process.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
