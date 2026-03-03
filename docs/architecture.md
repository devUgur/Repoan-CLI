# Architektur & Design 🏗️

Dieses Dokument beschreibt die interne Struktur und das Design-Prinzip von Repoan.

## 📁 Ordnerstruktur

Die Architektur folgt dem Standard-Layout für Go-CLI-Anwendungen:

- **`cmd/`**: CLI-Einstiegspunkt und Befehlsdefinitionen (Cobra).
- **`internal/`**: Kernfunktionalität, nicht exportierbar für externe Pakete.
    - **`scan/`**: Dateisystem-Scanner mit Unterstützung für Gitignore und Ignore-Muster.
    - **`model/`**: Gemeinsames Datenmodell (Snapshots, Findings).
    - **`analyze/`**: Logik für Sicherheits- und Hygiene-Checks (Regeln).
    - **`config/`**: Verwaltung der YAML-Konfiguration.
    - **`git/`**: Hilfsfunktionen für Git (z.B. Root-Erkennung).
    - **`output/`**: Formatter für verschiedene Ausgabeformate (Text, MD, JSON, SARIF).
    - **`logging/`**: Strukturiertes Logging via `slog`.

---

## 🛠️ Design-Prinzipien

### 1. Repository-Awareness
Repoan ist darauf ausgelegt, automatisch den Git-Root eines Repositories zu finden. Dies ermöglicht es, `repoan tree` von jedem Unterverzeichnis aus aufzurufen und dennoch das gesamte Repository zu sehen.

### 2. Plug-and-Play Analyzer
Analyse-Regeln sind über ein Interface definiert (`internal/analyze/analyzer.go`). Neue Regeln können einfach hinzugefügt werden, indem man eine neue Datei in `internal/analyze/` erstellt, die das Interface erfüllt.

### 3. Konfiguration vor Flags
Repoan nutzt eine "Configuration-first" Strategie. Alle Standardwerte werden in `.repoan/config.yml` definiert, können aber durch Flags für einmalige Aufrufe überschrieben werden.

### 4. CI/CD Fokus
Durch die Unterstützung des **SARIF-Formats** lässt sich Repoan nahtlos in GitHub Code Scanning integrieren. Ergebnisse können direkt in PRs angezeigt werden.

---

## 🚀 Technologie-Stack

- **Sprache:** Go (1.22+)
- **CLI Framework:** [Cobra](github.com/spf13/cobra)
- **Ignore Matching:** [go-gitignore](github.com/sabhiram/go-gitignore)
- **YAML Handling:** [yaml.v3](gopkg.in/yaml.v3)
- **Releasing:** [GoReleaser](.goreleaser.yaml)
