# Architecture & Design 🏗️

This document describes the internal structure and design principles of Repoan.

## 📁 Folder Structure

The architecture follows the standard layout for Go CLI applications:

- **`cmd/`**: CLI entry point and command definitions (Cobra).
- **`internal/`**: Core functionality, not exportable for external packages.
    - **`scan/`**: File system scanner with support for Gitignore and ignore patterns.
    - **`model/`**: Shared data model (Snapshots, Findings).
    - **`analyze/`**: Logic for security and hygiene checks (rules).
    - **`config/`**: Management of the YAML configuration.
    - **`git/`**: Helper functions for Git (e.g., root detection).
    - **`output/`**: Formatters for various output formats (Text, MD, JSON, SARIF).
    - **`logging/`**: Structured logging via `slog`.

---

## 🛠️ Design Principles

### 1. Repository Awareness
Repoan is designed to automatically find the Git root of a repository. This allows you to call `repoan tree` from any subdirectory and still see the entire repository.

### 2. Plug-and-Play Analyzer
Analysis rules are defined via an interface (`internal/analyze/analyzer.go`). New rules can be easily added by creating a new file in `internal/analyze/` that implements the interface.

### 3. Configuration Over Flags
Repoan follows a "Configuration-first" strategy. All default values are defined in `.repoan/config.yml`, but can be overridden by flags for one-time calls.

### 4. CI/CD Focus
By supporting the **SARIF format**, Repoan can be seamlessly integrated into GitHub Code Scanning. Results can be displayed directly in PRs.

---

## 🚀 Technology Stack

- **Language:** Go (1.22+)
- **CLI Framework:** [Cobra](github.com/spf13/cobra)
- **Ignore Matching:** [go-gitignore](github.com/sabhiram/go-gitignore)
- **YAML Handling:** [yaml.v3](gopkg.in/yaml.v3)
- **Releasing:** [GoReleaser](.goreleaser.yaml)
