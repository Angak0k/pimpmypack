package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Angak0k/pimpmypack/tests/cmd/apitest/internal/runner"
	"github.com/spf13/cobra"
)

var (
	runAll         bool
	scenarioDir    string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [scenario-numbers...]",
	Short: "Run API test scenarios",
	Long: `Run one or more API test scenarios by their number prefix.

Scenarios are YAML files in the tests/api-scenarios directory.
Use the numeric prefix to specify which scenarios to run.

Examples:
  apitest run 001              # Run scenario 001-*.yaml
  apitest run 001 002          # Run scenarios 001 and 002
  apitest run --all            # Run all scenarios in order`,
	RunE: runScenarios,
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVar(&runAll, "all", false, "Run all test scenarios")
	runCmd.Flags().StringVar(&scenarioDir, "scenario-dir", "tests/api-scenarios", "Directory containing test scenarios")
}

func runScenarios(cmd *cobra.Command, args []string) error {
	// Determine which scenarios to run
	var scenarioPaths []string

	if runAll {
		// Find all YAML files in scenario directory
		matches, err := filepath.Glob(filepath.Join(scenarioDir, "*.yaml"))
		if err != nil {
			return fmt.Errorf("failed to find scenarios: %w", err)
		}
		if len(matches) == 0 {
			return fmt.Errorf("no scenarios found in %s", scenarioDir)
		}
		sort.Strings(matches)
		scenarioPaths = matches
	} else if len(args) > 0 {
		// Run specific scenarios by number
		for _, num := range args {
			pattern := filepath.Join(scenarioDir, fmt.Sprintf("%s-*.yaml", num))
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return fmt.Errorf("failed to find scenario %s: %w", num, err)
			}
			if len(matches) == 0 {
				fmt.Printf("⚠️  Warning: No scenario found matching %s\n", num)
				continue
			}
			scenarioPaths = append(scenarioPaths, matches...)
		}
	} else {
		return fmt.Errorf("no scenarios specified (use scenario numbers or --all)")
	}

	if len(scenarioPaths) == 0 {
		return fmt.Errorf("no scenarios to run")
	}

	// Create test runner
	testRunner := runner.New(baseURL, verbose)

	// Check server health
	if err := testRunner.CheckServer(); err != nil {
		return err
	}

	fmt.Printf("\n")
	fmt.Printf("ℹ️  Running %d scenario(s)\n", len(scenarioPaths))

	overallStart := time.Now()

	// Run each scenario
	for _, scenarioPath := range scenarioPaths {
		if err := testRunner.Run(scenarioPath); err != nil {
			fmt.Printf("⚠️  Scenario error: %v\n", err)
			// Continue with other scenarios
		}
	}

	overallDuration := time.Since(overallStart)

	// Print summary
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📊 Test Summary")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	total, passed, failed := testRunner.GetSummary()
	fmt.Printf("Total tests:   %d\n", total)
	fmt.Printf("Passed tests:  \033[0;32m%d ✅\033[0m\n", passed)
	fmt.Printf("Failed tests:  \033[0;31m%d ❌\033[0m\n", failed)
	fmt.Printf("Duration:      %.2fs\n", overallDuration.Seconds())
	fmt.Println()

	if failed == 0 && total > 0 {
		fmt.Println("\033[0;32m✅ All tests passed! 🎉\033[0m")
		return nil
	} else if failed > 0 {
		fmt.Println("\033[0;31m❌ Some tests failed\033[0m")
		os.Exit(1)
	} else {
		fmt.Println("⚠️  No tests were run")
		os.Exit(1)
	}

	return nil
}

// getScenarioName extracts a friendly name from the scenario file path
func getScenarioName(path string) string {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, ".yaml")
	return name
}
