package oauth

import (
	"net/http"
	"os"

	"github.com/Vilamuzz/yota-backend/config"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func InitOAuth() {
	cfg := config.GetOAuthConfig()

	// Set session store with proper configuration
	key := []byte(cfg.SessionSecret)
	maxAge := 86400 * 30 // 30 days

	store := sessions.NewCookieStore(key)
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = os.Getenv("APP_ENV") == "production"
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	// Configure providers
	goth.UseProviders(
		google.New(
			cfg.GoogleClientID,
			cfg.GoogleClientSecret,
			cfg.GoogleCallbackURL,
			"email", "profile",
		),
	)
}
