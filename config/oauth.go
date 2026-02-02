package config

import "os"

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string
	SessionSecret      string
}

func GetOAuthConfig() OAuthConfig {
	return OAuthConfig{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleCallbackURL:  os.Getenv("GOOGLE_CALLBACK_URL"),
		SessionSecret:      os.Getenv("SESSION_SECRET"),
	}
}
