package config

import (
	"os"
	"strconv"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func GetSMTPConfig() SMTPConfig {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if port == 0 {
		port = 587 // default SMTP port
	}

	return SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     port,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
}
