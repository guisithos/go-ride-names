package main

import (
	"log"
	"os"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	clientID := os.Getenv("STRAVA_CLIENT_ID")
	clientSecret := os.Getenv("STRAVA_CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("STRAVA_CLIENT_ID and STRAVA_CLIENT_SECRET must be set in .env file")
	}

	auth.StartOAuthFlow(clientID, clientSecret)
}
