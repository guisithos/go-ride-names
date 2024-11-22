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
	// Print current working directory
	if dir, err := os.Getwd(); err == nil {
		fmt.Printf("Current working directory: %s\n", dir)
	}

	// Look for .env in project root (one level up from cmd)
	if err := godotenv.Load("../.env"); err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// Print environment variables for debugging
	fmt.Printf("STRAVA_CLIENT_ID: '%s'\n", os.Getenv("STRAVA_CLIENT_ID"))
	fmt.Printf("STRAVA_CLIENT_SECRET: '%s'\n", os.Getenv("STRAVA_CLIENT_SECRET"))

	config := &Config{
		StravaClientID:     os.Getenv("STRAVA_CLIENT_ID"),
		StravaClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
	}

	// Server config
	config.Server.Host = getEnvOrDefault("SERVER_HOST", "localhost")
	config.Server.Port = getEnvOrDefault("SERVER_PORT", "8080")

	// CORS config
	config.CORS.AllowedOrigins = strings.Split(getEnvOrDefault("CORS_ALLOWED_ORIGINS", "*"), ",")
	config.CORS.AllowedMethods = strings.Split(getEnvOrDefault("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ",")
	config.CORS.AllowedHeaders = strings.Split(getEnvOrDefault("CORS_ALLOWED_HEADERS", "Content-Type,Authorization"), ",")

	// OAuth config
	config.OAuth.RedirectURI = getEnvOrDefault("OAUTH_REDIRECT_URI", fmt.Sprintf("http://%s:%s/callback", config.Server.Host, config.Server.Port))

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
