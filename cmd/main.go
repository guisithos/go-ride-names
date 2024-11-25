package main

import (
	"log"
	"net/http"
	"os"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/handlers"
	"github.com/guisithos/go-ride-names/internal/service"
	"github.com/guisithos/go-ride-names/internal/strava"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create shared components with Redis
	redisURL := os.Getenv("REDIS_URL")
	sessions := auth.NewSessionStore(redisURL)
	mux := http.NewServeMux()

	// Setup OAuth handler with configured redirect URI
	oauthHandler := auth.NewOAuthHandler(cfg, sessions)
	oauthHandler.RegisterRoutes(mux)

	// Create webhook handler
	stravaClient := strava.NewClient("", "", cfg.StravaClientID, cfg.StravaClientSecret)
	activityService := service.NewActivityService(stravaClient)
	webhookHandler := handlers.NewWebhookHandler(activityService)
	webhookHandler.RegisterRoutes(mux)

	// Setup web handler
	webHandler := handlers.NewWebHandler(sessions, oauthHandler.GetConfig(), cfg)
	webHandler.RegisterRoutes(mux)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
