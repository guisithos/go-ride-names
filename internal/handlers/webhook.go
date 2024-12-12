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
			if err := h.processActivityWebhook(event); err != nil {
				log.Printf("Error processing webhook: %v", err)
				http.Error(w, "Error processing webhook", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *WebhookHandler) processActivityWebhook(event WebhookEvent) error {
	ownerID := fmt.Sprintf("%d", event.OwnerID)
	tokensInterface, exists := h.store.GetTokens(ownerID)
	if !exists {
		return fmt.Errorf("no tokens found for athlete %s", ownerID)
	}

	tokenData, err := json.Marshal(tokensInterface)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %v", err)
	}

	var tokens auth.TokenResponse
	if err := json.Unmarshal(tokenData, &tokens); err != nil {
		return fmt.Errorf("failed to unmarshal token data: %v", err)
	}

	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)
	activityService := service.NewActivityService(client)

	return activityService.RenameActivity(event.ObjectID)
}
