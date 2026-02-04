package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set during build via ldflags
	Version = "dev"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display version information for the PimpMyPack API test runner.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("PimpMyPack API Test Runner v%s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
