package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information set at build time via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("toolbox %s\n", Version)
		if Commit != "unknown" {
			fmt.Printf("  commit: %s\n", Commit)
		}
		if BuildDate != "unknown" {
			fmt.Printf("  built:  %s\n", BuildDate)
		}
	},
}
