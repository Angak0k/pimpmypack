package helper

import (
	"crypto/rand"
	"math/big"
	"net/smtp"
	"strconv"

	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/joho/godotenv"
)

func StringToUint(s string) (uint, error) {
	// Convert a string to an unsigned int.
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
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

func EnvInit() error {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}

func FinUserIDByUsername(users []dataset.User, username string) uint {
	// Find a user ID by username
	for _, user := range users {
		if user.Username == username {
			return user.ID
		}
	}
	return 0
}

func FinPackIDByPackName(packs dataset.Packs, packname string) uint {
	// Find a pack ID by packname
	for _, pack := range packs {
		if pack.Pack_name == packname {
			return pack.ID
		}
	}
	return 0
}

func FinItemIDByItemName(inventories dataset.Inventories, itemname string) uint {
	// Find an item ID by itemname
	for _, item := range inventories {
		if item.Item_name == itemname {
			return item.ID
		}
	}
	return 0
}

func GenerateRandomCode(length int) (string, error) {
	const charset = "pimpMyPackIsBetterThanLighterPack"
	var code string
	for i := 0; i < length; i++ {
		charIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code += string(charset[charIndex.Int64()])
	}
	return code, nil
}

// EmailSender defines the interface for sending emails.
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

	return smtp.SendMail(s.Server.MailServer+":"+strconv.Itoa(s.Server.MailPort), auth, s.Server.MailIdentity, []string{to}, msg)
}
