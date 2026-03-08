# CLI Reference 📖

Detailed overview of all available commands and flags in Repoan.

---

## 🔝 Global Flags
These flags can be used with any command:

- `--config <path>`: Path to the configuration file (default: `.repoan/config.yml`).
- `--debug`: Enable detailed debug logging (default: `false`).
- `-h, --help`: Show help for a command.
- `-v, --version`: Show version number.

---

## 📂 `repoan init`
Initializes Repoan in a new repository.

### Flags:
- `-f, --force`: Force overwrite existing configuration files.

### Effect:
Creates the `.repoan/` folder with `config.yml` and `baseline.json` and updates `.gitignore`.

---

## 🌳 `repoan tree [path]`
Scans the directory and outputs a tree structure.

### Arguments:
- `[path]`: Starting directory for the scan (default: Git root or current directory).

### Flags:
- `-d, --max-depth <int>`: Maximum depth of the scan.
- `--dirs-only`: Show directories only (omit files).
- `-f, --format <string>`: Output format: `txt`, `md`, `json`.
- `-i, --ignore <strings>`: List of ignore patterns.
- `-o, --out <path>`: Write the output to a file instead of the console.
- `--ascii`: Use ASCII tree characters (auto-enabled when output is redirected).
- `--here`: Scan starting from the current folder instead of searching for the Git root.
- `--respect-gitignore <bool>`: Whether to respect `.gitignore` files (default: `true`).

### Recommended usage for clean output:
- `repoan tree --dirs-only --max-depth 3`
- `repoan tree --dirs-only > structure.txt`
- `repoan tree --ignore .next/ --ignore coverage/ --ignore .turbo/ --dirs-only`

---

## 📸 `repoan snapshot [path]`
Creates a detailed JSON snapshot of the repository.

### Flags:
- `-o, --out <path>`: Filename for the snapshot (default: `repoan.snapshot.json`).
- `--here`: Scan starting from the current folder.

---

## 🔍 `repoan analyze [path]`
Performs security and hygiene checks.

### Flags:
- `-f, --format <string>`: Output format: `text`, `json`, `sarif`.
- `-o, --out <path>`: Write results to a file.
- `--fail-on <severity>`: Exit with code 1 if findings with this severity are found (`info`, `warning`, `high`, `critical`).

---

## 🧭 `repoan structure [path]`
Analyzes architecture structure with token patterns and module dependency metrics.

### Flags:
- `-f, --format <string>`: Output format: `text`, `md`, `json`, `sarif`, `mermaid`, `mermaid-cluster`.
- `-o, --out <path>`: Write structure report to a file.
- `--top-tokens <int>`: Number of top ranked structural tokens to emit.
- `--include-tests`: Include test files in dependency metric calculation.
- `--raw-dependencies`: Include raw file-level dependency traces in JSON output.
- `--raw-dependencies-limit <int>`: Limit raw dependency records (`0` = unlimited).
- `--raw-dependencies-only-unresolved`: Include only unresolved raw dependency records.
- `--raw-dependencies-only-external`: Include only external raw dependency records.
- `--raw-dependencies-min-confidence <float>`: Include only raw dependency records with confidence >= value (`0..1`).
- `--here`: Scan starting from the current folder instead of searching for the Git root.

### Example usage:
- `repoan structure`
- `repoan structure --format json --out structure-report.json`
- `repoan structure --format mermaid --out deps.mmd`
- `repoan structure --format mermaid-cluster --out deps-cluster.mmd`
- `repoan structure --format sarif --out structure.sarif`
- `repoan structure --format json --raw-dependencies --raw-dependencies-limit 500 --out structure-debug.json`
- `repoan structure --format json --raw-dependencies --raw-dependencies-only-unresolved --out unresolved-debug.json`
- `repoan structure --format json --raw-dependencies --raw-dependencies-only-external --raw-dependencies-min-confidence 0.4 --out external-low-confidence-debug.json`

---

## 🚪 Exit Codes
Repoan uses the following exit codes for automation:

| Code | Meaning |
| --- | --- |
| 0 | Success (no issues found) |
| 1 | Findings found above threshold (`--fail-on`) |
| 2 | General error |
| 3 | Usage error (wrong flags/arguments) |
| 4 | Filesystem error |
| 5 | No Git repository found |
