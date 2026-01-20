package helper

import (
	"crypto/rand"
	"math/big"
	"net/smtp"
	"regexp"
	"strconv"
	"strings"

	"github.com/Angak0k/pimpmypack/pkg/dataset"
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

func FindUserIDByUsername(users []dataset.User, username string) uint {
	// Find a user ID by username
	for _, user := range users {
		if user.Username == username {
			return user.ID
		}
	}
	return 0
}

func FindPackIDByPackName(packs dataset.Packs, packname string) uint {
	// Find a pack ID by packname
	for _, pack := range packs {
		if pack.PackName == packname {
			return pack.ID
		}
	}
	return 0
}

func FindItemIDByItemName(inventories dataset.Inventories, itemname string) uint {
	// Find an item ID by itemname
	for _, item := range inventories {
		if item.ItemName == itemname {
			return item.ID
		}
	}
	return 0
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
	SendEmail(to, subject, body string) error
}

// SMTPClient struct implements EmailSender interface.
type SMTPClient struct {
	Server dataset.MailServer
}

// SendMail sends an email using the SMTP protocol.
func (s *SMTPClient) SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.Server.MailUsername, s.Server.MailPassword, s.Server.MailServer)
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail(
		s.Server.MailServer+":"+strconv.Itoa(s.Server.MailPort),
		auth,
		s.Server.MailIdentity,
		[]string{to},
		msg,
	)
}
