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

// EmailSender defines the interface for sending emails. Needed for testing without real SMTP server.
type EmailSender interface {
	SendEmail(to, subject, textBody, htmlBody string) error
}

// SMTPClient struct implements EmailSender interface.
type SMTPClient struct {
	Server config.MailServer
}

// SendEmail sends an email using the SMTP protocol with proper MIME multipart/alternative formatting.
func (s *SMTPClient) SendEmail(to, subject, textBody, htmlBody string) error {
	auth := smtp.PlainAuth("", s.Server.MailUsername, s.Server.MailPassword, s.Server.MailServer)

	// Determine the sender address and display name.
	// MailIdentity may be an email (production) or a display name (.env.sample).
	// MailUsername may be a plain login or an email address.
	fromAddr := s.Server.MailIdentity
	fromName := "PimpMyPack"
	if !strings.Contains(fromAddr, "@") {
		// MailIdentity is a display name, fall back to MailUsername as address
		fromName = s.Server.MailIdentity
		fromAddr = s.Server.MailUsername
	}

	msg, err := BuildMIMEMessage(
		fromName, fromAddr,
		to, subject, textBody, htmlBody,
	)
	if err != nil {
		return fmt.Errorf("failed to build email message: %w", err)
	}

	return smtp.SendMail(
		s.Server.MailServer+":"+strconv.Itoa(s.Server.MailPort),
		auth,
		fromAddr,
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
		"MIME-Version: 1.0\r\n"+
		"Content-Type: multipart/alternative; boundary=%s\r\n"+
		"\r\n",
		fromName, fromAddr,
		to,
		subject,
		time.Now().Format(time.RFC1123Z),
		messageID,
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
