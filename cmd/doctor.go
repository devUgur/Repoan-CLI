package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/repoan/repoan/internal/git"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check the health of your repoan setup",
	Long:  `Performs self-diagnosis to ensure repoan is correctly configured in your repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🩺 Running repoan health check...")

		allOk := true

		// 1. Check Git Root
		root, err := git.FindRoot(".")
		if err != nil || root == "" {
			fmt.Println("❌ Not a git repository or git root not found.")
			allOk = false
		} else {
			fmt.Printf("✅ Git repository detected at: %s\n", root)
		}

		// 2. Check .repoan directory
		repoanDir := ".repoan"
		if info, err := os.Stat(repoanDir); err == nil && info.IsDir() {
			fmt.Println("✅ .repoan/ directory exists.")
		} else {
			fmt.Println("❌ .repoan/ directory missing. Run 'repoan init' to set it up.")
			allOk = false
		}

		// 3. Check config.yml
		cfgPath := filepath.Join(repoanDir, "config.yml")
		if _, err := os.Stat(cfgPath); err == nil {
			fmt.Println("✅ .repoan/config.yml found.")
		} else if allOk { // only warn if .repoan exists
			fmt.Println("⚠️  .repoan/config.yml missing. Using defaults.")
		}

		// 4. Check baseline.json
		baselinePath := filepath.Join(repoanDir, "baseline.json")
		if _, err := os.Stat(baselinePath); err == nil {
			fmt.Println("✅ .repoan/baseline.json found.")
		}

		// 5. Check reports directory
		reportsDir := filepath.Join(repoanDir, "reports")
		if info, err := os.Stat(reportsDir); err == nil && info.IsDir() {
			fmt.Println("✅ .repoan/reports/ directory exists.")
		}

		if allOk {
			fmt.Println("\n✨ Your repoan setup looks healthy!")
		} else {
			fmt.Println("\n⚠️  Some issues were found in your setup.")
			os.Exit(ExitGeneric)
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
