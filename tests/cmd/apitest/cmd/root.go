package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	baseURL string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "apitest",
	Short: "PimpMyPack API Test Runner",
	Long: `A CLI tool for running automated API test scenarios defined in YAML files.

Examples:
  apitest run 001              # Run scenario 001
  apitest run 001 002          # Run scenarios 001 and 002
  apitest run --all            # Run all scenarios
  apitest run --verbose 001    # Run with verbose output`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	setupCommands()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func setupCommands() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "http://localhost:8080/api", "Base URL of the API server")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Add subcommands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)

	// Setup run command flags
	runCmd.Flags().BoolVar(&runAll, "all", false, "Run all test scenarios")
	runCmd.Flags().StringVar(&scenarioDir, "scenario-dir", "tests/api-scenarios", "Directory containing test scenarios")
}
