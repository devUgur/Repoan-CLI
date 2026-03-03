# Beitragen zu Repoan-CLI 🤝

Vielen Dank für dein Interesse, Repoan zu verbessern! Jede Hilfe ist willkommen – von Bug-Reports bis hin zu neuen Features.

## 🛠️ Entwicklungseinrichtung

1. **Go installieren:** Version 1.22 oder höher wird empfohlen.
2. **Repository klonen:**
   ```bash
   git clone https://github.com/repoan/repoan.git
   cd repoan
   ```
3. **Abhängigkeiten installieren:**
   ```bash
   go mod tidy
   ```
4. **Bauen:**
   ```bash
   go build -o repoan .
   ```

## 🧪 Tests ausführen

Wir legen großen Wert auf Stabilität. Bitte stelle sicher, dass alle Tests bestehen:
```bash
go test -v ./...
```

## 📝 Pull Request Prozess

1. Erstelle einen **Feature-Branch** (`git checkout -b feature/mein-tolles-feature`).
2. Implementiere deine Änderungen und füge **Unit-Tests** hinzu.
3. Dokumentiere neue Funktionen im `README.md` oder in den `docs/`.
4. Sende einen **Pull Request** gegen den `main` Branch.

## 🐛 Issues melden

Wenn du einen Fehler findest oder einen Verbesserungsvorschlag hast, erstelle bitte ein [Issue](https://github.com/repoan/repoan/issues) mit:
- Einer klaren Beschreibung.
- Schritten zur Reproduktion (bei Fehlern).
- Erwartetes vs. tatsächliches Verhalten.

---

Vielen Dank, dass du Teil der Repoan-Community bist!
