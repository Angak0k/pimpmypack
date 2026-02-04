package runner

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Scenario represents a complete test scenario loaded from YAML
type Scenario struct {
	Name    string `yaml:"name"`
	BaseURL string `yaml:"base_url"`
	Steps   []Step `yaml:"scenarios"`
}

// Step represents a single test step within a scenario
type Step struct {
	Name       string            `yaml:"name"`
	Request    Request           `yaml:"request"`
	Assertions []Assertion       `yaml:"assertions"`
	Store      map[string]string `yaml:"store"`
}

// Request defines an HTTP request to be made
type Request struct {
	Method   string            `yaml:"method"`
	Endpoint string            `yaml:"endpoint"`
	Headers  map[string]string `yaml:"headers,omitempty"`
	Body     map[string]any    `yaml:"body,omitempty"`
	File     *FileUpload       `yaml:"file,omitempty"`
}

// FileUpload defines a file to be uploaded
type FileUpload struct {
	Field string `yaml:"field"` // Form field name
	Path  string `yaml:"path"`  // Path to the file
}

// Assertion defines a validation rule for the response
type Assertion struct {
	Type     string `yaml:"type"` // status_code, json_path
	Expected any    `yaml:"expected,omitempty"`
	Path     string `yaml:"path,omitempty"`
	Exists   bool   `yaml:"exists,omitempty"`
	Equals   string `yaml:"equals,omitempty"`
	Contains string `yaml:"contains,omitempty"`
	Empty    bool   `yaml:"empty,omitempty"` // Check if array/object is empty
}

// LoadScenario loads and parses a YAML scenario file
func LoadScenario(filename string) (*Scenario, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read scenario file: %w", err)
	}

	var scenario Scenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if scenario.Name == "" {
		return nil, errors.New("scenario name is required")
	}
	if scenario.BaseURL == "" {
		return nil, errors.New("base_url is required")
	}
	if len(scenario.Steps) == 0 {
		return nil, errors.New("scenario must have at least one step")
	}

	return &scenario, nil
}
