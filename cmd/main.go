package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/handlers"
)

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting application...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize session store
	sessions := auth.NewSessionStore()
	log.Printf("Session store initialized")

	// Create router
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

	// Configure server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("Server listening on %s", addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("Error starting server: %v", err)

	case <-shutdown:
		log.Println("Starting shutdown...")

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown did not complete in %v: %v", 15*time.Second, err)
			if err := server.Close(); err != nil {
				log.Printf("Error killing server: %v", err)
			}
		}
	}
}
