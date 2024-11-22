package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/handlers"
	"github.com/guisithos/go-ride-names/internal/middleware"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create shared components
	sessions := auth.NewSessionStore()
	mux := http.NewServeMux()

	// Setup OAuth handler with configured redirect URI
	oauthHandler := auth.NewOAuthHandler(cfg, sessions)
	oauthHandler.RegisterRoutes(mux)

	// Setup web handler
	webHandler := handlers.NewWebHandler(sessions, oauthHandler.GetConfig(), cfg)
	webHandler.RegisterRoutes(mux)

	// Apply middleware chain with configured CORS
	handler := middleware.Chain(
		mux,
		middleware.Recovery,
		middleware.Logger,
		middleware.Health("1.0.0", "development"),
		middleware.CORS(middleware.CORSConfig{
			AllowedOrigins: cfg.CORS.AllowedOrigins,
			AllowedMethods: cfg.CORS.AllowedMethods,
			AllowedHeaders: cfg.CORS.AllowedHeaders,
		}),
	)

	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on http://%s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, handler))
}
