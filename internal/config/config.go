package config

import (
	"fmt"
	"os"
	"strings"

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
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Remove .env loading in production
	if os.Getenv("ENVIRONMENT") != "production" {
		if err := godotenv.Load(); err != nil {
			fmt.Printf("Warning: Error loading .env file: %v\n", err)
		}
	}

	// Add support for Google Cloud's PORT environment variable
	port := getEnvOrDefault("PORT", "8080")

	config := &Config{
		StravaClientID:     os.Getenv("STRAVA_CLIENT_ID"),
		StravaClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
	}

	// Update server config
	config.Server.Host = getEnvOrDefault("SERVER_HOST", "0.0.0.0")
	config.Server.Port = port

	// CORS config
	config.CORS.AllowedOrigins = strings.Split(getEnvOrDefault("CORS_ALLOWED_ORIGINS", "*"), ",")
	config.CORS.AllowedMethods = strings.Split(getEnvOrDefault("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ",")
	config.CORS.AllowedHeaders = strings.Split(getEnvOrDefault("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"), ",")

	// OAuth config
	defaultRedirectURI := "https://go-ride-names-se2bdxecnq-uc.a.run.app/callback"
	config.OAuth.RedirectURI = getEnvOrDefault("OAUTH_REDIRECT_URI", defaultRedirectURI)

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
