package main

import (
	"log"

	"github.com/guisithos/go-ride-names/config"
	"github.com/guisithos/go-ride-names/service"
	"github.com/guisithos/go-ride-names/strava"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize config
	cfg := config.New()

	// Initialize Strava client
	stravaClient := strava.NewClient(cfg.StravaAccessToken)

	// Initialize activity service
	activityService := service.NewActivityService(stravaClient)

	// Get authenticated athlete
	athlete, err := activityService.GetAuthenticatedAthlete()
	if err != nil {
		log.Fatalf("Error getting authenticated athlete: %v", err)
	}

	log.Printf("Successfully authenticated as athlete: %s %s", athlete.FirstName, athlete.LastName)
}
