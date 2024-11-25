package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	StravaClientID     string
	StravaClientSecret string
	BaseURL            string
	OAuth              struct {
		RedirectURI string
	}
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{}

	// Load Strava configuration
	config.StravaClientID = os.Getenv("STRAVA_CLIENT_ID")
	config.StravaClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
	config.BaseURL = getEnvOrDefault("BASE_URL", "http://localhost:8080")

	// If OAUTH_REDIRECT_URI is not set, construct it from BASE_URL
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = config.BaseURL + "/callback"
	}
	config.OAuth.RedirectURI = redirectURI

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
