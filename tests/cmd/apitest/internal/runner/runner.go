package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Angak0k/pimpmypack/tests/cmd/apitest/internal/client"
	"github.com/Angak0k/pimpmypack/tests/cmd/apitest/internal/output"
)

// Runner executes test scenarios
type Runner struct {
	client    *client.HTTPClient
	state     *State
	formatter *output.Formatter
	verbose   bool

	TotalTests  int
	PassedTests int
	FailedTests int
}

// New creates a new test runner
func New(baseURL string, verbose bool) *Runner {
	return &Runner{
		client:    client.New(baseURL),
		state:     NewState(),
		formatter: output.New(verbose),
		verbose:   verbose,
	}
}

// CheckServer verifies the server is reachable
func (r *Runner) CheckServer() error {
	r.formatter.PrintInfo("Checking if server is running...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := r.client.CheckServer(ctx); err != nil {
		r.formatter.PrintFail("Server is not reachable")
		return fmt.Errorf("server check failed: %w", err)
	}

	r.formatter.PrintPass("Server is running")
	return nil
}

// Run executes a test scenario from a file
func (r *Runner) Run(scenarioFile string) error {
	// Load scenario
	scenario, err := LoadScenario(scenarioFile)
	if err != nil {
		return fmt.Errorf("failed to load scenario: %w", err)
	}

	r.formatter.PrintHeader(fmt.Sprintf("Scenario: %s", scenario.Name))

	startTime := time.Now()

	// Execute each step
	for i, step := range scenario.Steps {
		r.TotalTests++
		r.formatter.PrintStep(i+1, step.Name)

		statusCode, err := r.executeStep(scenario.BaseURL, step)
		if err != nil {
			r.formatter.PrintFail(err.Error())
			r.FailedTests++

			// Continue with other steps (don't abort entire scenario)
			continue
		}

		r.formatter.PrintPass(fmt.Sprintf("Status: %d", statusCode))
		r.PassedTests++
	}

	duration := time.Since(startTime)
	fmt.Printf("\n")
	r.formatter.PrintInfo(fmt.Sprintf("Scenario completed in %.2fs", duration.Seconds()))

	return nil
}

// executeStep runs a single test step and returns the HTTP status code
func (r *Runner) executeStep(baseURL string, step Step) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Substitute variables in request
	req := r.substituteRequest(step.Request)

	// 2. Make HTTP request (with file upload if specified)
	var resp *http.Response
	var body []byte
	var err error

	if req.File != nil {
		// File upload request
		resp, body, err = r.client.MakeRequestWithFile(
			ctx,
			req.Method,
			req.Endpoint,
			req.Headers,
			req.File.Path,
			req.File.Field,
			req.File.ContentType,
		)
	} else {
		// Regular JSON request
		resp, body, err = r.client.MakeRequest(ctx, req.Method, req.Endpoint, req.Headers, req.Body)
	}

	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	// 3. Validate assertions (with variable substitution)
	for _, assertion := range step.Assertions {
		// Substitute variables in assertion values
		substitutedAssertion := r.substituteAssertion(assertion)
		if err := ValidateAssertion(substitutedAssertion, resp, body); err != nil {
			return resp.StatusCode, err
		}
	}

	// 4. Store variables
	if err := r.storeVariables(step.Store, body, req.Body); err != nil {
		return resp.StatusCode, fmt.Errorf("failed to store variables: %w", err)
	}

	return resp.StatusCode, nil
}

// substituteRequest replaces variables in a request
func (r *Runner) substituteRequest(req Request) Request {
	return Request{
		Method:   req.Method,
		Endpoint: r.state.Substitute(req.Endpoint),
		Headers:  r.state.SubstituteStringMap(req.Headers),
		Body:     r.state.SubstituteMap(req.Body),
		File:     req.File, // Preserve file upload info
	}
}

// substituteAssertion replaces variables in assertion values
func (r *Runner) substituteAssertion(assertion Assertion) Assertion {
	return Assertion{
		Type:     assertion.Type,
		Expected: assertion.Expected,
		Path:     assertion.Path,
		Exists:   assertion.Exists,
		Equals:   r.state.Substitute(assertion.Equals),
		Contains: r.state.Substitute(assertion.Contains),
		Empty:    assertion.Empty,
	}
}

// storeVariables saves values from response or request for later use
func (r *Runner) storeVariables(store map[string]string, responseBody []byte, requestBody map[string]any) error {
	for varName, expression := range store {
		var value string

		// Check if expression references response
		if len(expression) > 11 && expression[:11] == "{{response." {
			// Extract from response: {{response.field}}
			path := expression[11 : len(expression)-2] // Remove {{response. and }}

			var data map[string]any
			if err := json.Unmarshal(responseBody, &data); err != nil {
				return fmt.Errorf("failed to parse response JSON: %w", err)
			}

			val := getJSONPath(data, path)
			if val != nil {
				value = fmt.Sprint(val)
			}
		} else if len(expression) > 15 && expression[:15] == "{{request.body." {
			// Extract from request body: {{request.body.field}}
			path := expression[15 : len(expression)-2] // Remove {{request.body. and }}
			val := getJSONPath(requestBody, path)
			if val != nil {
				value = fmt.Sprint(val)
			}
		} else {
			// Direct value
			value = expression
		}

		r.state.Set(varName, value)

		if r.verbose {
			r.formatter.PrintVerbose(fmt.Sprintf("Stored %s = %s", varName, value))
		}
	}

	return nil
}

// GetSummary returns test execution summary
func (r *Runner) GetSummary() (total, passed, failed int) {
	return r.TotalTests, r.PassedTests, r.FailedTests
}
