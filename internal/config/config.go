package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	StravaClientID     string
	StravaClientSecret string
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

	// Validate required fields
	if config.StravaClientID == "" {
		return nil, fmt.Errorf("STRAVA_CLIENT_ID is required")
	}
	if config.StravaClientSecret == "" {
		return nil, fmt.Errorf("STRAVA_CLIENT_SECRET is required")
	}

	return config, nil
}
