package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// MailServer contains SMTP configuration
type MailServer struct {
	MailServer   string `json:"mail_server"`
	MailPort     int    `json:"mail_port"`
	MailIdentity string `json:"mail_identity"`
	MailUsername string `json:"mail_username"`
	MailPassword string `json:"mail_password"`
}

type DBConfig struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
}

var (
	Scheme                           string
	HostName                         string
	DBHost                           string
	DBUser                           string
	DBPassword                       string
	DBName                           string
	DBPort                           int
	Stage                            string
	APISecret                        string
	TokenLifespan                    int
	AccessTokenMinutes               int
	RefreshTokenDays                 int
	RefreshTokenRememberMeDays       int
	RefreshTokenCleanupIntervalHours int
	MailServerConfig                 MailServer
)

type Config struct {
	Scheme                           string
	HostName                         string
	DBConfig                         DBConfig
	Stage                            string
	APISecret                        string
	TokenLifespan                    int
	AccessTokenMinutes               int
	RefreshTokenDays                 int
	RefreshTokenRememberMeDays       int
	RefreshTokenCleanupIntervalHours int
	MailServer                       MailServer
}

func EnvInit(envFilePath string) error {
	newConfig, err := initConfig(envFilePath)
	if err != nil {
		return err
	}

	Scheme = newConfig.Scheme
	HostName = newConfig.HostName
	DBHost = newConfig.DBConfig.DBHost
	DBUser = newConfig.DBConfig.DBUser
	DBPassword = newConfig.DBConfig.DBPassword
	DBName = newConfig.DBConfig.DBName
	DBPort = newConfig.DBConfig.DBPort
	Stage = newConfig.Stage
	APISecret = newConfig.APISecret
	TokenLifespan = newConfig.TokenLifespan
	AccessTokenMinutes = newConfig.AccessTokenMinutes
	RefreshTokenDays = newConfig.RefreshTokenDays
	RefreshTokenRememberMeDays = newConfig.RefreshTokenRememberMeDays
	RefreshTokenCleanupIntervalHours = newConfig.RefreshTokenCleanupIntervalHours
	MailServerConfig = newConfig.MailServer

	return nil
}

func initConfig(envFilePath string) (Config, error) {
	if err := loadEnv(envFilePath); err != nil {
		return Config{}, err
	}

	cfg := newConfig()
	setEnvVars(&cfg)

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// loadEnv loads the environment variables from a file if it exists.
func loadEnv(envFilePath string) error {
	if _, err := os.Stat(envFilePath); err == nil {
		if err := godotenv.Load(envFilePath); err != nil {
			return errors.New("error loading .env file")
		}
	}
	return nil
}

// newConfig returns a new Config struct with default values
func newConfig() Config {
	return Config{
		Scheme:                           "https",
		TokenLifespan:                    1,
		AccessTokenMinutes:               15,
		RefreshTokenDays:                 1,
		RefreshTokenRememberMeDays:       30,
		RefreshTokenCleanupIntervalHours: 24,
		APISecret:                        "defaultApiSecret",
		Stage:                            "local",
		HostName:                         "localhost",
		DBConfig: DBConfig{
			DBPort: 5432,
		},
		MailServer: MailServer{
			MailPort: 587,
		},
	}
}

func setEnvVars(cfg *Config) {
	cfg.Scheme = ifEnvEmpty(os.Getenv("SCHEME"), cfg.Scheme)
	cfg.HostName = ifEnvEmpty(os.Getenv("HOSTNAME"), cfg.HostName)
	cfg.DBConfig.DBHost = os.Getenv("DB_HOST")
	cfg.DBConfig.DBUser = os.Getenv("DB_USER")
	cfg.DBConfig.DBPassword = os.Getenv("DB_PASSWORD")
	cfg.DBConfig.DBName = os.Getenv("DB_NAME")
	cfg.DBConfig.DBPort, _ = strconv.Atoi(ifEnvEmpty(os.Getenv("DB_PORT"), strconv.Itoa(cfg.DBConfig.DBPort)))
	cfg.Stage = ifEnvEmpty(os.Getenv("STAGE"), cfg.Stage)
	cfg.APISecret = ifEnvEmpty(os.Getenv("API_SECRET"), cfg.APISecret)
	cfg.TokenLifespan, _ = strconv.Atoi(ifEnvEmpty(os.Getenv("TOKEN_HOUR_LIFESPAN"), strconv.Itoa(cfg.TokenLifespan)))
	cfg.AccessTokenMinutes, _ = strconv.Atoi(
		ifEnvEmpty(os.Getenv("ACCESS_TOKEN_MINUTES"), strconv.Itoa(cfg.AccessTokenMinutes)))
	cfg.RefreshTokenDays, _ = strconv.Atoi(
		ifEnvEmpty(os.Getenv("REFRESH_TOKEN_DAYS"), strconv.Itoa(cfg.RefreshTokenDays)))
	cfg.RefreshTokenRememberMeDays, _ = strconv.Atoi(
		ifEnvEmpty(os.Getenv("REFRESH_TOKEN_REMEMBER_ME_DAYS"), strconv.Itoa(cfg.RefreshTokenRememberMeDays)))
	cfg.RefreshTokenCleanupIntervalHours, _ = strconv.Atoi(
		ifEnvEmpty(os.Getenv("REFRESH_TOKEN_CLEANUP_INTERVAL_HOURS"), strconv.Itoa(cfg.RefreshTokenCleanupIntervalHours)))
	cfg.MailServer.MailIdentity = os.Getenv("MAIL_IDENTITY")
	cfg.MailServer.MailUsername = os.Getenv("MAIL_USERNAME")
	cfg.MailServer.MailPassword = os.Getenv("MAIL_PASSWORD")
	cfg.MailServer.MailServer = os.Getenv("MAIL_SERVER")
	cfg.MailServer.MailPort, _ = strconv.Atoi(ifEnvEmpty(os.Getenv("MAIL_PORT"), strconv.Itoa(cfg.MailServer.MailPort)))
}

func ifEnvEmpty(envVar, defaultValue string) string {
	if envVar == "" {
		return defaultValue
	}
	return envVar
}

func validateConfig(cfg Config) error {
	switch {
	case cfg.DBConfig.DBHost == "":
		return errors.New("DB_HOST is not set")
	case cfg.DBConfig.DBUser == "":
		return errors.New("DB_USER is not set")
	case cfg.DBConfig.DBPassword == "":
		return errors.New("DB_PASSWORD is not set")
	case cfg.DBConfig.DBName == "":
		return errors.New("DB_NAME is not set")
	case cfg.MailServer.MailIdentity == "":
		return errors.New("MAIL_IDENTITY is not set")
	case cfg.MailServer.MailUsername == "":
		return errors.New("MAIL_USERNAME is not set")
	case cfg.MailServer.MailPassword == "":
		return errors.New("MAIL_PASSWORD is not set")
	case cfg.MailServer.MailServer == "":
		return errors.New("MAIL_SERVER is not set")
	}
	return nil
}
