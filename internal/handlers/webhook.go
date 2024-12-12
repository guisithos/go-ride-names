package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/service"
	"github.com/guisithos/go-ride-names/internal/storage"
	"github.com/guisithos/go-ride-names/internal/strava"
)

type WebhookEvent struct {
	ObjectType string                 `json:"object_type"`
	ObjectID   int64                  `json:"object_id"`
	AspectType string                 `json:"aspect_type"`
	OwnerID    int64                  `json:"owner_id"`
	Updates    map[string]interface{} `json:"updates"`
}

type WebhookHandler struct {
	store        storage.Store
	stravaConfig *config.Config
	verifyToken  string
}

func NewWebhookHandler(store storage.Store, stravaConfig *config.Config) *WebhookHandler {
	verifyToken := os.Getenv("WEBHOOK_VERIFY_TOKEN")
	if verifyToken == "" {
		log.Println("Warning: WEBHOOK_VERIFY_TOKEN not set")
		// Generate a random token as fallback
		token := make([]byte, 32)
		rand.Read(token)
		verifyToken = hex.EncodeToString(token)
	}

	return &WebhookHandler{
		store:        store,
		stravaConfig: stravaConfig,
		verifyToken:  verifyToken,
	}
}

func (h *WebhookHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/webhook", h.handleWebhook)
}

func (h *WebhookHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received webhook request: Method=%s, URL=%s", r.Method, r.URL.String())

	switch r.Method {
	case http.MethodGet:
		// Handle Strava's webhook verification
		challenge := r.URL.Query().Get("hub.challenge")
		if challenge != "" {
			log.Printf("Received webhook verification challenge: %s", challenge)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"hub.challenge": challenge,
			})
			return
		}

	case http.MethodPost:
		var event WebhookEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			log.Printf("Error decoding webhook event: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Only process 'create' events for activities
		if event.ObjectType == "activity" && event.AspectType == "create" {
			log.Printf("Processing new activity: %d", event.ObjectID)

			// Get athlete's tokens
			ownerID := fmt.Sprintf("%d", event.OwnerID)
			tokensInterface, exists := h.store.GetTokens(ownerID)
			if !exists {
				log.Printf("No tokens found for athlete %s", ownerID)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokens, ok := tokensInterface.(*auth.TokenResponse)
			if !ok {
				log.Printf("Invalid token type for athlete %s", ownerID)
				http.Error(w, "Invalid token data", http.StatusInternalServerError)
				return
			}

			client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
				h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)
			activityService := service.NewActivityService(client)

			if err := activityService.RenameActivity(event.ObjectID); err != nil {
				log.Printf("Error renaming activity: %v", err)
				http.Error(w, "Error processing activity", http.StatusInternalServerError)
				return
			}

			log.Printf("Successfully processed activity %d", event.ObjectID)
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
