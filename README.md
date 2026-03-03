# Repoan-CLI 🚀

[![Go Report Card](https://goreportcard.com/badge/github.com/repoan/repoan)](https://goreportcard.com/report/github.com/repoan/repoan)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/repoan/repoan)](https://github.com/repoan/repoan/releases)

**Repoan** ist ein professionelles, schnelles und plattformübergreifendes CLI-Tool zur Snapshot-Erstellung von Verzeichnisstrukturen und zur Repository-Analyse. Es hilft Entwicklern und Teams dabei, Codebases zu dokumentieren, Repository-Kontexte für LLMs vorzubereiten und Sicherheits- sowie Hygiene-Checks in CI/CD-Pipelines zu integrieren.

> 💡 **Repoan ist leichtgewichtig, sprachagnostisch und offline-first.** Es dient als effizienter Pre-Scanner für lokale Entwicklung und CI/CD-Gates, ohne Vendor Lock-in oder SaaS-Zwang.

---

## ✨ Features

- **Repository-Aware Scanning:** Automatische Erkennung des Git-Roots.
- **Intelligentes Filtern:** Berücksichtigt `.gitignore` und benutzerdefinierte Ignore-Muster aus `.repoan/config.yml`.
- **Flexible Formate:** Ausgabe der Verzeichnisstruktur als Text, Markdown oder JSON.
- **Repository Snapshots:** Erstellung detaillierter JSON-Snapshots inklusive Dateigrößen und Metadaten.
- **Integrierte Analyse:**
    - **Security:** Erkennt sensible Dateien (z.B. `.env`, `.pem`, Keys).
    - **Hygiene:** Identifiziert zu große Dateien oder Binaries.
- **CI/CD Ready:** Exportiert Ergebnisse im **SARIF-Format** (nativ unterstützt von GitHub Code Scanning) und bietet `--fail-on` Schwellenwerte.
- **Zentralisierte Konfiguration:** Alles sauber in einem `.repoan/` Verzeichnis organisiert.
- **Deterministisch:** Identische Repositories erzeugen identische Ergebnisse – perfekt für Stable-CI.

---

## 🚀 Installation

### 1. Via Go (für Entwickler)
```powershell
go install github.com/repoan/repoan@latest
```
Stelle sicher, dass dein `GOPATH/bin` in deinem System-PATH enthalten ist.

### 2. Binaries (Windows, macOS, Linux)
Lade das passende Binary für dein System von der [Releases-Seite](https://github.com/repoan/repoan/releases) herunter und füge es zu deinem PATH hinzu.

---

## 🛠️ Erste Schritte

### Initialisierung
Bereite dein Repository für Repoan vor. Dies erstellt einen `.repoan/` Ordner mit Standardeinstellungen.
```powershell
repoan init
```

### Verzeichnisstruktur anzeigen (Tree)
```powershell
# Standard Baumansicht (respektiert .gitignore)
repoan tree

# Als Markdown für Dokumentationen
repoan tree --format md

# Begrenzte Tiefe
repoan tree --max-depth 2
```

### Repository Snapshot erstellen
```powershell
repoan snapshot --out repo-snapshot.json
```

### Analyse ausführen
```powershell
# Schnell-Check (Textausgabe)
repoan analyze

# Export für GitHub Code Scanning (SARIF)
repoan analyze --format sarif --out results.sarif
```

---

## ⚙️ Konfiguration (`.repoan/config.yml`)

Repoan lässt sich über eine zentrale YAML-Datei steuern. Hier ein Beispiel:

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
  fail_on: "high"
output:
  dir: ".repoan/reports"
  default_format: "md"
```

---

## 🤖 CI/CD Integration

Repoan ist für den Einsatz in Pipelines optimiert. Beispiel für GitHub Actions:

```yaml
- name: Run Repoan Analysis
  run: repoan analyze --format sarif --out repoan-results.sarif --fail-on high

- name: Upload SARIF results
  uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: repoan-results.sarif
```

---

## 🤝 Beitragen

Wir freuen uns über Beiträge! Schau in unsere [CONTRIBUTING.md](CONTRIBUTING.md) für Details zum Entwicklungsprozess.

## 📄 Lizenz

Dieses Projekt ist unter der MIT-Lizenz lizenziert - siehe die [LICENSE](LICENSE) Datei für Details.
