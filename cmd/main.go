package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/service"
	"github.com/guisithos/go-ride-names/internal/strava"

	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env file from different possible locations
	envPaths := []string{
		".env",                       // Current directory
		"../.env",                    // Parent directory
		filepath.Join("cmd", ".env"), // cmd directory
	}

	envLoaded := false
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			envLoaded = true
			break
		}
	}

	if !envLoaded {
		log.Println("Warning: .env file not found. Make sure environment variables are set.")
	}

	// Verify that we have the required environment variable
	if os.Getenv("STRAVA_ACCESS_TOKEN") == "" {
		log.Fatal("STRAVA_ACCESS_TOKEN environment variable is not set")
	}

	// Initialize config
	cfg := config.New()

	// Initialize Strava client
	stravaClient := strava.NewClient(
		cfg.StravaAccessToken,
		cfg.StravaRefreshToken,
		cfg.StravaClientID,
		cfg.StravaClientSecret,
	)

	// Initialize service
	activityService := service.NewActivityService(stravaClient)

	// Get authenticated athlete
	athlete, err := activityService.GetAuthenticatedAthlete()
	if err != nil {
		log.Fatalf("Error getting authenticated athlete: %v", err)
	}

	log.Printf("Successfully authenticated as athlete: %s %s", athlete.FirstName, athlete.LastName)

	// Get activities for the last 30 days
	now := time.Now().Unix()
	thirtyDaysAgo := now - (30 * 24 * 60 * 60)

	activities, err := activityService.ListActivities(1, 30, now, thirtyDaysAgo, true)
	if err != nil {
		log.Fatalf("Error getting activities: %v", err)
	}

	// Print activities
	fmt.Printf("Found %d activities:\n", len(activities))
	for _, activity := range activities {
		fmt.Printf("- %s (%.2f km) on %s\n",
			activity.Name,
			activity.Distance/1000, // Convert meters to kilometers
			activity.StartDateLocal.Format("2006-01-02 15:04:05"),
		)
	}
}
