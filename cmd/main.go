package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/handlers"
	"github.com/guisithos/go-ride-names/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Configure Google Cloud Storage
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("GCS_BUCKET_NAME environment variable is required")
	}

	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	ctx := context.Background()

	// Initialize storage
	store, err := storage.NewGCSStore(ctx, bucketName, credentialsFile)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize handlers
	mux := http.NewServeMux()

	// Setup OAuth handler
	oauthHandler := auth.NewOAuthHandler(cfg, store)
	oauthHandler.RegisterRoutes(mux)

	// Create webhook handler
	webhookHandler := handlers.NewWebhookHandler(store, cfg)
	webhookHandler.RegisterRoutes(mux)

	// Setup web handler
	webHandler := handlers.NewWebHandler(store, oauthHandler.GetConfig(), cfg)
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
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
