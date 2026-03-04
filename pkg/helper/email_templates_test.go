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
		{"confirm URL href", `href="https://example.com/confirm?id=1&code=abc"`},
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
