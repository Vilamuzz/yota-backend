package main

import (
	"log"
	"os"

	"github.com/Vilamuzz/yota-backend/internal/app"
	"github.com/Vilamuzz/yota-backend/pkg/oauth"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize OAuth
	oauth.InitOAuth()

	application, cleanup, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}
	defer cleanup()

	if err := application.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
