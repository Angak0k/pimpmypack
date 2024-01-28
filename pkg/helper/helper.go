package helper

import (
	"net/smtp"
	"strconv"

	"github.com/Angak0k/pimpmypack/pkg/config"
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

func SendEmail(to, subject, body string) error {

	auth := smtp.PlainAuth("", config.MailUsername, config.MailPassword, config.MailServer)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	return smtp.SendMail(config.MailServer+":"+strconv.Itoa(config.MailPort), auth, config.MailIdentity, []string{to}, msg)

}
