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
- `-f, --format <string>`: Output format: `txt`, `md`, `json`.
- `-i, --ignore <strings>`: List of ignore patterns.
- `-o, --out <path>`: Write the output to a file instead of the console.
- `--here`: Scan starting from the current folder instead of searching for the Git root.
- `--respect-gitignore <bool>`: Whether to respect `.gitignore` files (default: `true`).

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
