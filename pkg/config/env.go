package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var (
	DbHost        string
	DbUser        string
	DbPassword    string
	DbName        string
	DbPort        string
	Stage         string
	ApiSecret     string
	TokenLifespan int
	MailIdentity  string
	MailUsername  string
	MailPassword  string
	MailServer    string
	MailPort      int
)

func EnvInit(envFilePath string) error {
	var err error

	if _, err := os.Stat(envFilePath); err == nil {
		err := godotenv.Load(envFilePath)
		if err != nil {
			return fmt.Errorf("error loading .env file")
		}
	}

	DbHost = os.Getenv("DB_HOST")
	DbUser = os.Getenv("DB_USER")
	DbPassword = os.Getenv("DB_PASSWORD")
	DbName = os.Getenv("DB_NAME")
	DbPort = os.Getenv("DB_PORT")
	Stage = os.Getenv("STAGE")
	ApiSecret = os.Getenv("API_SECRET")
	TokenLifespan, err = strconv.Atoi(os.Getenv("TOKEN_HOUR_LIFESPAN"))
	if err != nil {
		TokenLifespan = 1
	}
	MailIdentity = os.Getenv("MAIL_IDENTITY")
	MailUsername = os.Getenv("MAIL_USERNAME")
	MailPassword = os.Getenv("MAIL_PASSWORD")
	MailServer = os.Getenv("MAIL_SERVER")
	MailPort, err = strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		return err
	}
	return nil
}
