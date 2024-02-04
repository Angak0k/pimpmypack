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
		log.Fatalf("Error loading .env file or environement variable : %v", err)
	}

	ret := m.Run()
	os.Exit(ret)

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
	code, err := GenerateRandomCode(10)
	if err != nil {
		t.Errorf("GenerateRandomCode failed: %v", err)
	}
	if len(code) != 10 {
		t.Errorf("Expected code length of 10, got %d", len(code))
	}
}
