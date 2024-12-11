package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
	case "GET":
		// Handle webhook validation
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		log.Printf("Webhook validation request - Mode: %s, Token: %s, Challenge: %s",
			mode, token, challenge)

		if mode != "subscribe" || token != h.verifyToken {
			log.Printf("Invalid webhook validation - Expected token: %s, Got: %s",
				h.verifyToken, token)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		log.Printf("Webhook validation successful")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"hub.challenge": challenge,
		})

	case "POST":
		// Handle webhook events
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading webhook body: %v", err)
			http.Error(w, "Error reading request", http.StatusBadRequest)
			return
		}
		log.Printf("Received webhook event: %s", string(body))

		var event WebhookEvent
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("Error parsing webhook event: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Process only new activities
		if event.ObjectType == "activity" && event.AspectType == "create" {
			log.Printf("Processing new activity: ID=%d", event.ObjectID)

			// Get tokens from session
			ownerID := fmt.Sprintf("%d", event.OwnerID)
			tokens, exists := h.store.GetTokens(ownerID)
			if !exists {
				log.Printf("No tokens found for athlete %s", ownerID)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenResp, ok := tokens.(*auth.TokenResponse)
			if !ok {
				log.Printf("Invalid token type for athlete %s", ownerID)
				http.Error(w, "Invalid token data", http.StatusInternalServerError)
				return
			}

			client := strava.NewClient(tokenResp.AccessToken, tokenResp.RefreshToken,
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
