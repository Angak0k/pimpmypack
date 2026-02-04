package runner

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ValidateAssertion checks if a response matches the assertion criteria
func ValidateAssertion(assertion Assertion, resp *http.Response, body []byte) error {
	switch assertion.Type {
	case "status_code":
		return validateStatusCode(assertion, resp)
	case "json_path":
		return validateJSONPath(assertion, body)
	default:
		return fmt.Errorf("unknown assertion type: %s", assertion.Type)
	}
}

// validateStatusCode checks the HTTP status code
func validateStatusCode(assertion Assertion, resp *http.Response) error {
	var expected int

	// Handle Expected as different types (int, float64, string)
	switch v := assertion.Expected.(type) {
	case int:
		expected = v
	case float64:
		expected = int(v)
	case string:
		var err error
		expected, err = strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid expected status code: %v", assertion.Expected)
		}
	default:
		return fmt.Errorf("expected status code must be a number, got %T", assertion.Expected)
	}

	if resp.StatusCode != expected {
		return fmt.Errorf("expected status %d, got %d", expected, resp.StatusCode)
	}

	return nil
}

// validateJSONPath checks JSON response fields
func validateJSONPath(assertion Assertion, body []byte) error {
	// Try to unmarshal as generic interface to support both objects and arrays
	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Extract value using JSON path
	value := getJSONPath(data, assertion.Path)

	// Check if field should exist
	if assertion.Exists {
		if value == nil {
			return fmt.Errorf("field %s does not exist", assertion.Path)
		}
		return nil // Field exists, assertion passes
	}

	// Check if collection is empty
	if assertion.Empty {
		isEmpty := false
		switch v := value.(type) {
		case []any:
			isEmpty = len(v) == 0
		case map[string]any:
			isEmpty = len(v) == 0
		case nil:
			isEmpty = true
		default:
			return fmt.Errorf("field %s: cannot check empty on type %T", assertion.Path, value)
		}

		if !isEmpty {
			return fmt.Errorf("field %s: expected empty collection, but has %d elements", assertion.Path, getLength(value))
		}
		return nil
	}

	// Check for exact match
	if assertion.Equals != "" {
		actual := fmt.Sprint(value)
		if actual != assertion.Equals {
			return fmt.Errorf("field %s: expected '%s', got '%s'", assertion.Path, assertion.Equals, actual)
		}
		return nil
	}

	// Check for substring match
	if assertion.Contains != "" {
		actual := fmt.Sprint(value)
		if !strings.Contains(actual, assertion.Contains) {
			return fmt.Errorf("field %s: expected to contain '%s', got '%s'", assertion.Path, assertion.Contains, actual)
		}
		return nil
	}

	return nil
}

// getJSONPath extracts a value from JSON using a path like "$.field.nested" or "$[0].field"
func getJSONPath(data any, path string) any {
	// Remove "$" prefix if present
	path = strings.TrimPrefix(path, "$")

	// Handle empty path
	if path == "" || path == "." {
		return data
	}

	// Remove leading "." if present
	path = strings.TrimPrefix(path, ".")

	var current any = data
	i := 0

	for i < len(path) {
		// Handle array index [n]
		if path[i] == '[' {
			// Find closing bracket
			closeBracket := strings.Index(path[i:], "]")
			if closeBracket == -1 {
				return nil
			}

			// Extract index
			indexStr := path[i+1 : i+closeBracket]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil
			}

			// Access array element
			arr, ok := current.([]any)
			if !ok {
				return nil
			}
			if index < 0 || index >= len(arr) {
				return nil
			}
			current = arr[index]

			// Move past the bracket
			i += closeBracket + 1

			// Skip the dot after bracket if present
			if i < len(path) && path[i] == '.' {
				i++
			}
			continue
		}

		// Handle object field
		// Find next delimiter (. or [)
		nextDelim := len(path)
		for j := i; j < len(path); j++ {
			if path[j] == '.' || path[j] == '[' {
				nextDelim = j
				break
			}
		}

		key := path[i:nextDelim]
		if key != "" {
			m, ok := current.(map[string]any)
			if !ok {
				return nil
			}
			current = m[key]
			if current == nil {
				return nil
			}
		}

		i = nextDelim
		// Skip the dot
		if i < len(path) && path[i] == '.' {
			i++
		}
	}

	return current
}

// getLength returns the length of a collection (array or object)
func getLength(value any) int {
	switch v := value.(type) {
	case []any:
		return len(v)
	case map[string]any:
		return len(v)
	default:
		return 0
	}
}
