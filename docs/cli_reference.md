# CLI Referenz 📖

Detaillierte Übersicht aller verfügbaren Befehle und Flags in Repoan.

---

## 🔝 Globale Flags
Diese Flags können mit jedem Befehl verwendet werden:

- `--config <pfad>`: Pfad zur Konfigurationsdatei (Standard: `.repoan/config.yml`).
- `--debug`: Aktiviert detaillierte Debug-Logs (Standard: `false`).
- `-h, --help`: Hilfe zu einem Befehl anzeigen.
- `-v, --version`: Versionsnummer anzeigen.

---

## 📂 `repoan init`
Initialisiert Repoan in einem neuen Repository.

### Flags:
- `-f, --force`: Erzwingt das Überschreiben bestehender Konfigurationsdateien.

### Effekt:
Erstellt den `.repoan/` Ordner mit `config.yml` und `baseline.json` und aktualisiert die `.gitignore`.

---

## 🌳 `repoan tree [pfad]`
Scannt das Verzeichnis und gibt eine Baumstruktur aus.

### Argumente:
- `[pfad]`: Startverzeichnis für den Scan (Standard: Git-Root oder aktuelles Verzeichnis).

### Flags:
- `-d, --max-depth <int>`: Maximale Tiefe des Scans.
- `-f, --format <string>`: Ausgabeformat: `txt`, `md`, `json`.
- `-i, --ignore <strings>`: Liste von Ignore-Mustern.
- `-o, --out <pfad>`: Schreibt die Ausgabe in eine Datei statt in die Konsole.
- `--here`: Scannt ab dem aktuellen Ordner, statt den Git-Root zu suchen.
- `--respect-gitignore <bool>`: Ob `.gitignore` Dateien beachtet werden sollen (Standard: `true`).

---

## 📸 `repoan snapshot [pfad]`
Erstellt einen detaillierten JSON-Snapshot des Repositories.

### Flags:
- `-o, --out <pfad>`: Dateiname für den Snapshot (Standard: `repoan.snapshot.json`).
- `--here`: Scannt ab dem aktuellen Ordner.

---

## 🔍 `repoan analyze [pfad]`
Führt Sicherheits- und Hygiene-Checks durch.

### Flags:
- `-f, --format <string>`: Ausgabeformat: `text`, `json`, `sarif`.
- `-o, --out <pfad>`: Schreibt Ergebnisse in eine Datei.
- `--fail-on <severity>`: Bricht den Befehl mit Exit-Code 1 ab, wenn Findings mit dieser Severity gefunden werden (`info`, `warning`, `high`, `critical`).
