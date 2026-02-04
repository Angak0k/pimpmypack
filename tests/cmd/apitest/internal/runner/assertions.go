package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ValidateAssertion checks if a response matches the assertion criteria
func ValidateAssertion(assertion Assertion, statusCode int, body []byte) error {
	switch assertion.Type {
	case "status_code":
		return validateStatusCode(assertion, statusCode)
	case "json_path":
		return validateJSONPath(assertion, body)
	default:
		return fmt.Errorf("unknown assertion type: %s", assertion.Type)
	}
}

// validateStatusCode checks the HTTP status code
func validateStatusCode(assertion Assertion, actualStatus int) error {
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

	if actualStatus != expected {
		return fmt.Errorf("expected status %d, got %d", expected, actualStatus)
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
		var isEmpty bool
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
	// Remove "$" prefix and leading "." if present
	path = strings.TrimPrefix(strings.TrimPrefix(path, "$"), ".")

	// Handle empty path
	if path == "" {
		return data
	}

	current := data
	i := 0

	for i < len(path) {
		if path[i] == '[' {
			var err error
			current, i, err = processArrayIndex(current, path, i)
			if err != nil {
				return nil
			}
			continue
		}

		var err error
		current, i, err = processObjectField(current, path, i)
		if err != nil {
			return nil
		}
	}

	return current
}

// processArrayIndex handles array indexing in JSON path
func processArrayIndex(current any, path string, pos int) (any, int, error) {
	// Find closing bracket
	closeBracket := strings.Index(path[pos:], "]")
	if closeBracket == -1 {
		return nil, 0, errors.New("missing closing bracket")
	}

	// Extract and parse index
	indexStr := path[pos+1 : pos+closeBracket]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return nil, 0, err
	}

	// Access array element
	arr, ok := current.([]any)
	if !ok || index < 0 || index >= len(arr) {
		return nil, 0, errors.New("invalid array access")
	}

	newPos := pos + closeBracket + 1
	// Skip the dot after bracket if present
	if newPos < len(path) && path[newPos] == '.' {
		newPos++
	}

	return arr[index], newPos, nil
}

// processObjectField handles object field access in JSON path
func processObjectField(current any, path string, pos int) (any, int, error) {
	// Find next delimiter (. or [)
	nextDelim := len(path)
	for j := pos; j < len(path); j++ {
		if path[j] == '.' || path[j] == '[' {
			nextDelim = j
			break
		}
	}

	key := path[pos:nextDelim]
	if key != "" {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, 0, errors.New("not an object")
		}
		current = m[key]
		if current == nil {
			return nil, 0, errors.New("field not found")
		}
	}

	newPos := nextDelim
	// Skip the dot
	if newPos < len(path) && path[newPos] == '.' {
		newPos++
	}

	return current, newPos, nil
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
