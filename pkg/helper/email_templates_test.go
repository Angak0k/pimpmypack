package helper

import (
	"strings"
	"testing"
)

func TestBuildConfirmationEmailHTML(t *testing.T) {
	html := BuildConfirmationEmailHTML("testuser", "https://example.com/confirm?id=1&code=abc")

	checks := []struct {
		name   string
		substr string
	}{
		{"username", "testuser"},
		{"confirm URL href", `href="https://example.com/confirm?id=1&amp;code=abc"`},
		{"CTA button text", "Confirm my email"},
		{"branding", "PimpMyPack"},
		{"brand header color", "#14532d"},
		{"button color", "#16a34a"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(html, c.substr) {
				t.Errorf("expected HTML to contain %q", c.substr)
			}
		})
	}
}

func TestBuildConfirmationEmailText(t *testing.T) {
	text := BuildConfirmationEmailText("testuser", "https://example.com/confirm?id=1&code=abc")

	checks := []struct {
		name   string
		substr string
	}{
		{"username", "testuser"},
		{"confirm URL", "https://example.com/confirm?id=1&code=abc"},
		{"branding", "PimpMyPack"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(text, c.substr) {
				t.Errorf("expected text to contain %q", c.substr)
			}
		})
	}
}

func TestBuildPasswordResetEmailHTML(t *testing.T) {
	html := BuildPasswordResetEmailHTML("s3cretP@ss")

	checks := []struct {
		name   string
		substr string
	}{
		{"password", "s3cretP@ss"},
		{"warning", "did not request"},
		{"branding", "PimpMyPack"},
		{"recommendation", "recommend changing"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(html, c.substr) {
				t.Errorf("expected HTML to contain %q", c.substr)
			}
		})
	}
}

func TestBuildConfirmationEmailText_UsernameInjection(t *testing.T) {
	malicious := "alice\r\n\r\nInjected body content"
	text := BuildConfirmationEmailText(malicious, "https://example.com/confirm")

	// CRLF must be stripped so the username appears as a single token
	if strings.Contains(text, "alice\r\n") || strings.Contains(text, "alice\n") {
		t.Error("username with CRLF was not sanitized in plain-text email")
	}
	// After sanitization the username line should read "Welcome, aliceInjected body content!"
	// (harmless concatenation, no extra lines injected)
	if !strings.Contains(text, "Welcome, aliceInjected body content!") {
		t.Error("expected sanitized username to be concatenated without newlines")
	}
}

func TestBuildConfirmationEmailHTML_UsernameInjection(t *testing.T) {
	malicious := "<script>alert('xss')</script>"
	html := BuildConfirmationEmailHTML(malicious, "https://example.com/confirm")

	if strings.Contains(html, "<script>") {
		t.Error("HTML body must escape script tags in username")
	}
}

func TestBuildPasswordResetEmailText(t *testing.T) {
	text := BuildPasswordResetEmailText("s3cretP@ss")

	checks := []struct {
		name   string
		substr string
	}{
		{"password", "s3cretP@ss"},
		{"warning", "did not request"},
		{"branding", "PimpMyPack"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(text, c.substr) {
				t.Errorf("expected text to contain %q", c.substr)
			}
		})
	}
}
