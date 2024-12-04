package config

import (
	"fmt"
	"log"
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
	log.Printf("=== Loading Configuration ===")

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}

	config := &Config{}

	// Load Strava configuration
	config.StravaClientID = os.Getenv("STRAVA_CLIENT_ID")
	config.StravaClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
	config.BaseURL = getEnvOrDefault("BASE_URL", "http://localhost:8080")

	log.Printf("Base URL: %s", config.BaseURL)
	log.Printf("Strava Client ID: %s", config.StravaClientID)
	log.Printf("Strava Client Secret: %s[redacted]", config.StravaClientSecret[:5])

	// If OAUTH_REDIRECT_URI is not set, construct it from BASE_URL
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = config.BaseURL + "/callback"
		log.Printf("No OAUTH_REDIRECT_URI set, using constructed value: %s", redirectURI)
	} else {
		log.Printf("Using configured OAUTH_REDIRECT_URI: %s", redirectURI)
	}
	config.OAuth.RedirectURI = redirectURI

	// Validate required fields
	if config.StravaClientID == "" {
		return nil, fmt.Errorf("STRAVA_CLIENT_ID is required")
	}
	if config.StravaClientSecret == "" {
		return nil, fmt.Errorf("STRAVA_CLIENT_SECRET is required")
	}

	log.Printf("=== Configuration Loaded Successfully ===")
	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
