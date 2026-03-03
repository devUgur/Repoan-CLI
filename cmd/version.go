package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of repoan",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("repoan version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
