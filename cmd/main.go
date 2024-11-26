package main

import (
	"log"
	"net/http"
	"os"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/handlers"
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

	// Create webhook handler with session-aware client
	webhookHandler := handlers.NewWebhookHandler(sessions, cfg)
	webhookHandler.RegisterRoutes(mux)

	// Setup web handler
	webHandler := handlers.NewWebHandler(sessions, oauthHandler.GetConfig(), cfg)
	webHandler.RegisterRoutes(mux)

	// Add static file serving
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
