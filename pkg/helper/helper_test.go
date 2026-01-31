package helper

import (
	"log"
	"os"
	"testing"

	"github.com/Angak0k/pimpmypack/pkg/config"
)

func TestMain(m *testing.M) {
	// init env
	err := config.EnvInit("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file or environment variable : %v", err)
	}

	ret := m.Run()
	os.Exit(ret)
}

func TestStringToUint(t *testing.T) {
	// Test a valid string
	val, err := StringToUint("10")
	if err != nil {
		t.Errorf("StringToUint failed: %v", err)
	}
	if val != 10 {
		t.Errorf("Expected 10, got %d", val)
	}

	// Test an invalid string
	_, err = StringToUint("invalid")
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

func TestConvertWeightUnit(t *testing.T) {
	// Define test cases
	testCases := []struct {
		input    string // Input weight unit
		expected string // Expected output
	}{
		{"gram", "METRIC"},
		{"oz", "IMPERIAL"},
		{"dog", "METRIC"}, // Assuming "dog" defaults to METRIC as per your original test logic
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			output := ConvertWeightUnit(tc.input)
			if output != tc.expected {
				t.Errorf("ConvertWeightUnit(%s): expected %s, got %s", tc.input, tc.expected, output)
			}
		})
	}
}

// MockEmailSender is a mock implementation of EmailSender for testing.
type MockEmailSender struct {
	SentEmails []Email // Store sent emails for verification
}

// Email represents an email message for testing.
type Email struct {
	To      string
	Subject string
	Body    string
}

// SendMail records the email sending action without actually sending an email.
func (m *MockEmailSender) SendEmail(to, subject, body string) error {
	m.SentEmails = append(m.SentEmails, Email{To: to, Subject: subject, Body: body})
	return nil // Return nil to simulate a successful send
}
func TestSendEmail(t *testing.T) {
	// Create a new instance of the mock
	mockSender := &MockEmailSender{}

	err := mockSender.SendEmail("example@example.com", "Test Subject", "This is a test.")
	if err != nil {
		t.Errorf("SendMail failed: %v", err)
	}

	// Verify that the email was "sent"
	if len(mockSender.SentEmails) != 1 {
		t.Errorf("Expected 1 email to be sent, got %d", len(mockSender.SentEmails))
	}
}

func TestGenerateRandomCode(t *testing.T) {
	// Define test cases in a slice of structs
	testCases := []struct {
		name     string // Name of the test case for readability
		length   int    // Input length for GenerateRandomCode
		expected int    // Expected length of the generated code
	}{
		{"Length 10", 10, 10},
		{"Length 0", 0, 0},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, err := GenerateRandomCode(tc.length)
			if err != nil {
				t.Errorf("GenerateRandomCode failed: %v", err)
			}
			if len(code) != tc.expected {
				t.Errorf("Expected code length of %d, got %d", tc.expected, len(code))
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	testCases := []struct {
		name     string // Name of the test case for readability
		email    string // Input email
		expected bool   // Expected output
	}{
		{"standard email", "test@exemple.com", true},
		{"no tld", "test@exemple", false},
		{"no @", "testexemple.com", false},
		{"no tld, dot terminated", "test@exemple.", false},
		{"composed tld", "test@exemple.co.test", true},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := IsValidEmail(tc.email)
			if output != tc.expected {
				t.Errorf("IsValidEmail(%s): expected %t, got %t", tc.email, tc.expected, output)
			}
		})
	}
}
