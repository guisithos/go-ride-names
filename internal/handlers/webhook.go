package handlers

import (
	"bytes"
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
	// Log full request details including body for all methods
	log.Printf("Webhook received: Method=%s, URL=%s, Headers=%v",
		r.Method, r.URL.String(), r.Header)

	// Read and log the body for all requests
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
	} else if len(body) > 0 {
		log.Printf("Request body: %s", string(body))
	}
	// Restore the body for further processing
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	switch r.Method {
	case http.MethodGet:
		// Handle Strava's webhook verification
		challenge := r.URL.Query().Get("hub.challenge")
		verifyToken := r.URL.Query().Get("hub.verify_token")

		log.Printf("Webhook verification - Challenge: %s, Token: %s, Expected Token: %s",
			challenge, verifyToken, h.verifyToken)

		if verifyToken != h.verifyToken {
			log.Printf("Invalid verify_token received: %s", verifyToken)
			http.Error(w, "Invalid verification token", http.StatusBadRequest)
			return
		}

		if challenge != "" {
			log.Printf("Responding to webhook verification challenge: %s", challenge)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"hub.challenge": challenge,
			})
			return
		}
		http.Error(w, "Invalid verification request", http.StatusBadRequest)
		return

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading webhook body: %v", err)
			http.Error(w, "Error reading request", http.StatusBadRequest)
			return
		}
		// Log the raw webhook payload
		log.Printf("Received webhook payload: %s", string(body))
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		var event WebhookEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			log.Printf("Error decoding webhook event: %v", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Log the decoded event
		log.Printf("Processing webhook event: Type=%s, ID=%d, AspectType=%s, OwnerID=%d",
			event.ObjectType, event.ObjectID, event.AspectType, event.OwnerID)

		if event.ObjectType == "activity" && event.AspectType == "create" {
			if err := h.processActivityWebhook(event); err != nil {
				log.Printf("Error processing webhook: %v", err)
				http.Error(w, "Error processing webhook", http.StatusInternalServerError)
				return
			}
			log.Printf("Successfully processed webhook for activity %d", event.ObjectID)
		} else {
			log.Printf("Skipping event: not a new activity (Type=%s, Aspect=%s)",
				event.ObjectType, event.AspectType)
		}

		w.WriteHeader(http.StatusOK)
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *WebhookHandler) processActivityWebhook(event WebhookEvent) error {
	log.Printf("Starting to process activity webhook for ID=%d", event.ObjectID)

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

	log.Printf("Attempting to rename activity %d", event.ObjectID)
	if err := activityService.RenameActivity(event.ObjectID); err != nil {
		return fmt.Errorf("failed to rename activity: %v", err)
	}

	log.Printf("Successfully renamed activity %d", event.ObjectID)
	return nil
}
