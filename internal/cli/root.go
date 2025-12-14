package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "toolbox",
	Short: "AI-assisted software development tools",
	Long:  `A collection of tools to assist in AI-enabled software development.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
