package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/joho/godotenv"
)

func main() {
	// Print current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %v", err)
	}
	log.Printf("Current working directory: %s", currentDir)

	// Try to load .env file from different possible locations
	envPaths := []string{
		".env",                       // Current directory
		"../.env",                    // Parent directory
		filepath.Join("cmd", ".env"), // cmd directory
		"../../.env",                 // Two levels up
	}

	envLoaded := false
	log.Println("Checking for .env file in:")
	for _, path := range envPaths {
		absPath, _ := filepath.Abs(path)
		log.Printf("- %s", absPath)
		if err := godotenv.Load(path); err == nil {
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Fatal("Error: .env file not found. Please create a .env file with STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET")
	}

	clientID := os.Getenv("STRAVA_CLIENT_ID")
	clientSecret := os.Getenv("STRAVA_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET must be set in .env file")
	}

	auth.StartOAuthFlow(clientID, clientSecret)
}
