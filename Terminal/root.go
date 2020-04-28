package Terminal

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ProgImage",
	Short: "ProgImage is an Convertors api",
	Long: `ProgImage is an Convertors api, created as an interview code task.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
