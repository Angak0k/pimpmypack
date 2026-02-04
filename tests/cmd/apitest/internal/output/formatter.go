package output

import (
	"fmt"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
)

// Formatter handles colored console output
type Formatter struct {
	verbose bool
}

// New creates a new output formatter
func New(verbose bool) *Formatter {
	return &Formatter{
		verbose: verbose,
	}
}

// PrintPass prints a success message
func (f *Formatter) PrintPass(message string) {
	fmt.Printf("%s✅ PASSED%s - %s\n", colorGreen, colorReset, message)
}

// PrintFail prints a failure message
func (f *Formatter) PrintFail(message string) {
	fmt.Printf("%s❌ FAILED%s - %s\n", colorRed, colorReset, message)
}

// PrintInfo prints an informational message
func (f *Formatter) PrintInfo(message string) {
	fmt.Printf("%sℹ️  %s%s\n", colorCyan, message, colorReset)
}

// PrintWarning prints a warning message
func (f *Formatter) PrintWarning(message string) {
	fmt.Printf("%s⚠️  %s%s\n", colorYellow, message, colorReset)
}

// PrintHeader prints a section header with borders
func (f *Formatter) PrintHeader(title string) {
	fmt.Printf("\n%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorBlue, colorReset)
	fmt.Printf("%s%s%s\n", colorBlue, title, colorReset)
	fmt.Printf("%s━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n\n", colorBlue, colorReset)
}

// PrintStep prints a step description
func (f *Formatter) PrintStep(stepNumber int, description string) {
	fmt.Printf("Step %d: %s\n", stepNumber, description)
}

// PrintSummary prints the final test summary
func (f *Formatter) PrintSummary(total, passed, failed int, duration string) {
	f.PrintHeader("📊 Test Summary")

	fmt.Printf("Total tests:   %d\n", total)
	fmt.Printf("Passed tests:  %s%d ✅%s\n", colorGreen, passed, colorReset)
	fmt.Printf("Failed tests:  %s%d ❌%s\n", colorRed, failed, colorReset)

	if duration != "" {
		fmt.Printf("Duration:      %s\n", duration)
	}

	fmt.Println()

	if failed == 0 {
		fmt.Printf("%s🎉 All tests passed!%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("%s⚠️  Some tests failed%s\n", colorYellow, colorReset)
	}
}

// PrintVerbose prints verbose debug information
func (f *Formatter) PrintVerbose(message string) {
	if f.verbose {
		fmt.Printf("%s[DEBUG]%s %s\n", colorCyan, colorReset, message)
	}
}

// PrintError prints an error message with details
func (f *Formatter) PrintError(err error) {
	fmt.Printf("%sError: %v%s\n", colorRed, err, colorReset)
}
