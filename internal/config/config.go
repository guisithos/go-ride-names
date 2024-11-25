package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	StravaClientID     string
	StravaClientSecret string

	Server struct {
		Host string
		Port string
	}
	CORS struct {
		AllowedOrigins []string
		AllowedMethods []string
		AllowedHeaders []string
	}
	OAuth struct {
		RedirectURI string
	}
	Redis struct {
		URL string
	}
	WebhookVerifyToken string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{}

	// Load Strava configuration
	config.StravaClientID = os.Getenv("STRAVA_CLIENT_ID")
	config.StravaClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
	config.OAuth.RedirectURI = getEnvOrDefault("OAUTH_REDIRECT_URI", "http://localhost:8080/callback")
	config.Redis.URL = getEnvOrDefault("REDIS_URL", "redis://localhost:6379")
	config.WebhookVerifyToken = getEnvOrDefault("WEBHOOK_VERIFY_TOKEN", "")

	// Validate required fields
	if config.StravaClientID == "" {
		return nil, fmt.Errorf("STRAVA_CLIENT_ID is required")
	}
	if config.StravaClientSecret == "" {
		return nil, fmt.Errorf("STRAVA_CLIENT_SECRET is required")
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
