package helper

import (
	"log"
	"os"
	"testing"

	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/dataset"
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

func TestFindUserIDByUsername(t *testing.T) {
	// Define test cases
	testCases := []struct {
		users []dataset.User
		// Input users
		username string // Input username
		expected uint   // Expected output
	}{
		{
			[]dataset.User{{ID: 1, Username: "user1"}, {ID: 2, Username: "user2"}},
			"user1",
			1,
		},
		{
			[]dataset.User{{ID: 1, Username: "user1"}, {ID: 2, Username: "user2"}},
			"user2",
			2,
		},
		{
			[]dataset.User{{ID: 1, Username: "user1"}, {ID: 2, Username: "user2"}},
			"user3",
			0,
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.username, func(t *testing.T) {
			output := FindUserIDByUsername(tc.users, tc.username)
			if output != tc.expected {
				t.Errorf("FinUserIDByUsername(%v, %s): expected %d, got %d", tc.users, tc.username, tc.expected, output)
			}
		})
	}
}

func TestFindPackIDByPackName(t *testing.T) {
	// Define test cases
	testCases := []struct {
		packs    dataset.Packs
		packname string // Input packname
		expected uint   // Expected output
	}{
		{
			dataset.Packs{{ID: 1, PackName: "pack1"}, {ID: 2, PackName: "pack2"}},
			"pack1",
			1,
		},
		{
			dataset.Packs{{ID: 1, PackName: "pack1"}, {ID: 2, PackName: "pack2"}},
			"pack2",
			2,
		},
		{
			dataset.Packs{{ID: 1, PackName: "pack1"}, {ID: 2, PackName: "pack2"}},
			"pack3",
			0,
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.packname, func(t *testing.T) {
			output := FindPackIDByPackName(tc.packs, tc.packname)
			if output != tc.expected {
				t.Errorf("FinPackIDByPackName(%v, %s): expected %d, got %d", tc.packs, tc.packname, tc.expected, output)
			}
		})
	}
}

func TestFindItemIDByItemName(t *testing.T) {
	// Define test cases
	testCases := []struct {
		inventories dataset.Inventories
		itemname    string // Input itemname
		expected    uint   // Expected output
	}{
		{
			dataset.Inventories{{ID: 1, ItemName: "item1"}, {ID: 2, ItemName: "item2"}},
			"item1",
			1,
		},
		{
			dataset.Inventories{{ID: 1, ItemName: "item1"}, {ID: 2, ItemName: "item2"}},
			"item2",
			2,
		},
		{
			dataset.Inventories{{ID: 1, ItemName: "item1"}, {ID: 2, ItemName: "item2"}},
			"item3",
			0,
		},
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.itemname, func(t *testing.T) {
			output := FindItemIDByItemName(tc.inventories, tc.itemname)
			if output != tc.expected {
				t.Errorf("FinItemIDByItemName(%v, %s): expected %d, got %d",
					tc.inventories,
					tc.itemname,
					tc.expected,
					output)
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
