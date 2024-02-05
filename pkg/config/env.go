package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Angak0k/pimpmypack/pkg/dataset"
	"github.com/joho/godotenv"
)

var (
	Scheme        string
	HostName      string
	DbHost        string
	DbUser        string
	DbPassword    string
	DbName        string
	DbPort        string
	Stage         string
	ApiSecret     string
	TokenLifespan int
	MailServer    dataset.MailServer
)

func EnvInit(envFilePath string) error {
	var err error

	if _, err := os.Stat(envFilePath); err == nil {
		err := godotenv.Load(envFilePath)
		if err != nil {
			return fmt.Errorf("error loading .env file")
		}
	}

	Scheme = os.Getenv("SCHEME")
	HostName = os.Getenv("HOSTNAME")
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
	MailServer.MailIdentity = os.Getenv("MAIL_IDENTITY")
	MailServer.MailUsername = os.Getenv("MAIL_USERNAME")
	MailServer.MailPassword = os.Getenv("MAIL_PASSWORD")
	MailServer.MailServer = os.Getenv("MAIL_SERVER")
	if os.Getenv("MAIL_PORT") != "" {
		MailServer.MailPort, err = strconv.Atoi(os.Getenv("MAIL_PORT"))
		if err != nil {
			return err
		}
	}

	// validate configuration
	switch {
	case Scheme != "http" && Scheme != "https":
		Scheme = "https"
	case HostName == "":
		return fmt.Errorf("HOSTNAME is not set and needed for generationg links")
	case DbHost == "":
		return fmt.Errorf("DB_HOST is not set")
	case DbUser == "":
		return fmt.Errorf("DB_USER is not set")
	case DbPassword == "":
		return fmt.Errorf("DB_PASSWORD is not set")
	case DbName == "":
		return fmt.Errorf("DB_NAME is not set")
	case DbPort == "":
		DbPort = "5432"
	case Stage == "":
		Stage = "prod"
	case ApiSecret == "":
		ApiSecret = "defaultApiSecret"
	case TokenLifespan == 0:
		TokenLifespan = 1
	case MailServer.MailIdentity == "":
		return fmt.Errorf("MAIL_IDENTITY is not set")
	case MailServer.MailUsername == "":
		return fmt.Errorf("MAIL_USERNAME is not set")
	case MailServer.MailPassword == "":
		return fmt.Errorf("MAIL_PASSWORD is not set")
	case MailServer.MailServer == "":
		return fmt.Errorf("MAIL_SERVER is not set")
	case MailServer.MailPort == 0:
		MailServer.MailPort = 587
	}

	return nil
}
