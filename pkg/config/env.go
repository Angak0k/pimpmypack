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
	DBHost        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBPort        int
	Stage         string
	APISecret     string
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
	DBHost = os.Getenv("DB_HOST")
	DBUser = os.Getenv("DB_USER")
	DBPassword = os.Getenv("DB_PASSWORD")
	DBName = os.Getenv("DB_NAME")
	DBPort, err = strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		DBPort = 5432
	}
	Stage = os.Getenv("STAGE")
	APISecret = os.Getenv("API_SECRET")
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
	case DBHost == "":
		return fmt.Errorf("DB_HOST is not set")
	case DBUser == "":
		return fmt.Errorf("DB_USER is not set")
	case DBPassword == "":
		return fmt.Errorf("DB_PASSWORD is not set")
	case DBName == "":
		return fmt.Errorf("DB_NAME is not set")
	case DBPort == 0:
		DBPort = 5432
	case Stage == "":
		Stage = "prod"
	case APISecret == "":
		APISecret = "defaultApiSecret"
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
