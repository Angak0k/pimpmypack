package helper

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Angak0k/pimpmypack/pkg/config"
)

func StringToUint(s string) (uint, error) {
	// Convert a string to an unsigned int.
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

func ComparePtrString(a, b *string) bool {
	// Compare two string pointers for equality
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ConvertWeightUnit(unit string) string {
	// Convert a weight unit to an enum, METRIC by default
	switch unit {
	case "gram":
		return "METRIC"
	case "oz":
		return "IMPERIAL"
	}
	return "METRIC"
}

func GenerateRandomCode(length int) (string, error) {
	const charset = "pimpMyPackIsBetterThanLighterPack"
	var builder strings.Builder
	for i := 0; i < length; i++ {
		charIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		builder.WriteByte(charset[charIndex.Int64()])
	}
	return builder.String(), nil
}

func IsValidEmail(email string) bool {
	// Check if an email is valid
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidUsername checks that a username does not contain '@' to avoid ambiguity with email login.
func IsValidUsername(username string) bool {
	return !strings.Contains(username, "@")
}

// EmailSender defines the interface for sending emails. Needed for testing without real SMTP server.
type EmailSender interface {
	SendEmail(to, subject, textBody, htmlBody string) error
}

// SMTPClient struct implements EmailSender interface.
type SMTPClient struct {
	Server config.MailServer
}

// SendEmail sends an email using the SMTP protocol with proper MIME multipart/alternative formatting.
// MailIdentity must be a valid email address (validated at startup) — used as envelope sender and From address.
// MailUsername is the SMTP authentication login credential.
func (s *SMTPClient) SendEmail(to, subject, textBody, htmlBody string) error {
	if !IsValidEmail(to) {
		return fmt.Errorf("invalid recipient email address: %s", to)
	}

	auth := smtp.PlainAuth("", s.Server.MailUsername, s.Server.MailPassword, s.Server.MailServer)

	msg, err := BuildMIMEMessage(
		"PimpMyPack", s.Server.MailIdentity,
		to, subject, textBody, htmlBody,
	)
	if err != nil {
		return fmt.Errorf("failed to build email message: %w", err)
	}

	return smtp.SendMail(
		s.Server.MailServer+":"+strconv.Itoa(s.Server.MailPort),
		auth,
		s.Server.MailIdentity,
		[]string{to},
		msg,
	)
}

// BuildMIMEMessage constructs a multipart/alternative MIME email message with proper headers.
func BuildMIMEMessage(
	fromName, fromAddr, to, subject, textBody, htmlBody string,
) ([]byte, error) {
	domain := extractDomain(fromAddr)
	messageID := generateMessageID(domain)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	headers := fmt.Sprintf("From: %s <%s>\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Date: %s\r\n"+
		"Message-ID: <%s>\r\n"+
		"List-Unsubscribe: <mailto:%s?subject=unsubscribe>\r\n"+
		"List-Unsubscribe-Post: List-Unsubscribe=One-Click\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: multipart/alternative; boundary=%s\r\n"+
		"\r\n",
		sanitizeHeaderValue(fromName), sanitizeHeaderValue(fromAddr),
		sanitizeHeaderValue(to),
		sanitizeHeaderValue(subject),
		time.Now().Format(time.RFC1123Z),
		messageID,
		sanitizeHeaderValue(fromAddr),
		writer.Boundary(),
	)

	var msg bytes.Buffer
	msg.WriteString(headers)

	// text/plain part
	textHeader := make(textproto.MIMEHeader)
	textHeader.Set("Content-Type", "text/plain; charset=utf-8")
	textHeader.Set("Content-Transfer-Encoding", "8bit")

	textPart, err := writer.CreatePart(textHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to create text part: %w", err)
	}
	if _, err := textPart.Write([]byte(textBody)); err != nil {
		return nil, fmt.Errorf("failed to write text part: %w", err)
	}

	// text/html part
	htmlHeader := make(textproto.MIMEHeader)
	htmlHeader.Set("Content-Type", "text/html; charset=utf-8")
	htmlHeader.Set("Content-Transfer-Encoding", "8bit")

	htmlPart, err := writer.CreatePart(htmlHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to create html part: %w", err)
	}
	if _, err := htmlPart.Write([]byte(htmlBody)); err != nil {
		return nil, fmt.Errorf("failed to write html part: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	msg.Write(buf.Bytes())

	return msg.Bytes(), nil
}

// sanitizeHeaderValue strips \r and \n characters to prevent SMTP header injection.
func sanitizeHeaderValue(v string) string {
	v = strings.ReplaceAll(v, "\r", "")
	v = strings.ReplaceAll(v, "\n", "")
	return v
}

func extractDomain(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return "localhost"
}

func generateMessageID(domain string) string {
	ts := time.Now().UnixNano()
	randBytes := make([]byte, 8)
	if _, err := rand.Read(randBytes); err != nil {
		return fmt.Sprintf("%d@%s", ts, domain)
	}
	return fmt.Sprintf("%d.%x@%s", ts, randBytes, domain)
}
