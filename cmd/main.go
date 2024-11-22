package main

import (
	"log"
	"net/http"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/handlers"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create shared components
	sessions := auth.NewSessionStore()
	mux := http.NewServeMux()

	// Setup OAuth handler
	oauthHandler := auth.NewOAuthHandler(cfg, sessions)
	oauthHandler.RegisterRoutes(mux)

	// Setup web handler
	webHandler := handlers.NewWebHandler(sessions, oauthHandler.GetConfig(), cfg)
	webHandler.RegisterRoutes(mux)

	log.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
