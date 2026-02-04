package runner

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// State manages variables that can be substituted in requests
type State struct {
	vars map[string]string
	mu   sync.RWMutex
}

// NewState creates a new state manager
func NewState() *State {
	return &State{
		vars: make(map[string]string),
	}
}

// Set stores a variable value
func (s *State) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars[key] = value
}

// Get retrieves a variable value
func (s *State) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.vars[key]
	return val, ok
}

// Substitute replaces {{variable}} placeholders in a string
func (s *State) Substitute(text string) string {
	// Handle built-in variables
	if strings.Contains(text, "{{timestamp}}") {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		text = strings.ReplaceAll(text, "{{timestamp}}", timestamp)
	}

	// Handle stored variables
	s.mu.RLock()
	defer s.mu.RUnlock()

	for key, value := range s.vars {
		placeholder := fmt.Sprintf("{{%s}}", key)
		text = strings.ReplaceAll(text, placeholder, value)
	}

	return text
}

// SubstituteMap replaces variables in a map[string]any (for request bodies)
func (s *State) SubstituteMap(m map[string]any) map[string]any {
	result := make(map[string]any)

	for key, value := range m {
		switch v := value.(type) {
		case string:
			substituted := s.Substitute(v)
			// Try to convert to number if it looks like a number
			result[key] = tryParseNumber(substituted)
		case map[string]any:
			result[key] = s.SubstituteMap(v)
		default:
			result[key] = value
		}
	}

	return result
}

// tryParseNumber attempts to parse a string as a number, returns original if not numeric
func tryParseNumber(s string) any {
	// Try parsing as integer first
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	// Try parsing as float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	// Try parsing as boolean
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	// Return original string if not a number
	return s
}

// SubstituteStringMap replaces variables in a map[string]string (for headers)
func (s *State) SubstituteStringMap(m map[string]string) map[string]string {
	result := make(map[string]string)

	for key, value := range m {
		result[key] = s.Substitute(value)
	}

	return result
}
