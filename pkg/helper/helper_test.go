package helper

import (
	"log"
	"os"
	"strings"
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
	To       string
	Subject  string
	TextBody string
	HTMLBody string
}

// SendEmail records the email sending action without actually sending an email.
func (m *MockEmailSender) SendEmail(to, subject, textBody, htmlBody string) error {
	m.SentEmails = append(m.SentEmails, Email{To: to, Subject: subject, TextBody: textBody, HTMLBody: htmlBody})
	return nil // Return nil to simulate a successful send
}

func TestSendEmail(t *testing.T) {
	// Create a new instance of the mock
	mockSender := &MockEmailSender{}

	err := mockSender.SendEmail("example@example.com", "Test Subject", "This is a test.", "<p>This is a test.</p>")
	if err != nil {
		t.Errorf("SendEmail failed: %v", err)
	}

	// Verify that the email was "sent"
	if len(mockSender.SentEmails) != 1 {
		t.Errorf("Expected 1 email to be sent, got %d", len(mockSender.SentEmails))
	}

	sent := mockSender.SentEmails[0]
	if sent.TextBody != "This is a test." {
		t.Errorf("Expected text body 'This is a test.', got '%s'", sent.TextBody)
	}
	if sent.HTMLBody != "<p>This is a test.</p>" {
		t.Errorf("Expected HTML body '<p>This is a test.</p>', got '%s'", sent.HTMLBody)
	}
}

func TestBuildMIMEMessage(t *testing.T) {
	msg, err := BuildMIMEMessage(
		"PimpMyPack", "noreply@pimpmypack.com",
		"user@example.com", "Test Subject",
		"Plain text body", "<p>HTML body</p>",
	)
	if err != nil {
		t.Fatalf("BuildMIMEMessage failed: %v", err)
	}

	raw := string(msg)

	checks := []struct {
		name   string
		substr string
	}{
		{"From header", "From: PimpMyPack <noreply@pimpmypack.com>"},
		{"To header", "To: user@example.com"},
		{"Subject header", "Subject: Test Subject"},
		{"Date header", "Date: "},
		{"Message-ID header", "Message-ID: <"},
		{"MIME-Version", "MIME-Version: 1.0"},
		{"multipart boundary", "Content-Type: multipart/alternative; boundary="},
		{"text/plain part", "Content-Type: text/plain; charset=utf-8"},
		{"text/html part", "Content-Type: text/html; charset=utf-8"},
		{"text body", "Plain text body"},
		{"html body", "<p>HTML body</p>"},
		{"8bit encoding", "Content-Transfer-Encoding: 8bit"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(raw, c.substr) {
				t.Errorf("expected message to contain %q", c.substr)
			}
		})
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

func TestIsValidUsername(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		expected bool
	}{
		{"simple username", "johndoe", true},
		{"username with numbers", "john123", true},
		{"username with hyphens", "john-doe", true},
		{"username with underscores", "john_doe", true},
		{"email-like username", "john@example.com", false},
		{"username with @ in middle", "john@doe", false},
		{"username starting with @", "@johndoe", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := IsValidUsername(tc.username)
			if output != tc.expected {
				t.Errorf("IsValidUsername(%s): expected %t, got %t", tc.username, tc.expected, output)
			}
		})
	}
}
