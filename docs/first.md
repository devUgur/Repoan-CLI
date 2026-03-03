Alles klar — hier ist der “5-Minuten”-Weg zu einem echten CLI mit `repoan --help` (Go + Cobra), inkl. Subcommand-Skelett.

## 1) Voraussetzungen checken

```bash
go version
git --version
```

## 2) Neues Projekt anlegen

```bash
mkdir repoan && cd repoan
go mod init github.com/<DEIN_USER>/repoan
```

## 3) Cobra installieren

```bash
go get github.com/spf13/cobra@latest
```

## 4) Minimal funktionierendes CLI erstellen

**main.go**

```go
package main

import "github.com/<DEIN_USER>/repoan/cmd"

func main() {
	cmd.Execute()
}
```

**cmd/root.go**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "repoan",
	Short: "repoan is a repository analytics CLI",
	Long:  "repoan is an open-source CLI to snapshot folder structures and analyze repositories.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

Jetzt Ordner anlegen:

```bash
mkdir -p cmd
```

Dann die Dateien speichern.

## 5) Subcommand hinzufügen: `tree` (Dummy)

**cmd/tree.go**

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Write the folder structure snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("tree: (TODO) will write structure.txt from current directory")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(treeCmd)
}
```

## 6) Bauen & testen

```bash
go build -o repoan .
./repoan --help
./repoan tree
./repoan tree --help
```

Du solltest jetzt sehen:

* `repoan --help` zeigt Usage + Commands
* `repoan tree` läuft und gibt die Dummy-Zeile aus

---

## Bonus: “wie installiert man das in PATH” (lokal)

Während der Entwicklung:

```bash
go install .
```

Dann liegt das Binary in:

* macOS/Linux: `~/go/bin/repoan` (meistens)
* Windows: `%USERPROFILE%\go\bin\repoan.exe`

Falls `~/go/bin` noch nicht im PATH ist, füge es hinzu (z. B. in `~/.zshrc`):

```bash
export PATH="$PATH:$HOME/go/bin"
```

Danach geht:

```bash
repoan --help
```

---

Wenn du willst, mache ich als nächsten Schritt direkt aus `repoan tree` eine echte Version mit:

* Snapshot ab aktuellem Pfad
* `--out structure.txt`
* Default ignore (`.git`, `node_modules`, …)
* optional `--format md|txt|json`
