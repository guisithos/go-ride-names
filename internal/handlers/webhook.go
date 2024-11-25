package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/guisithos/go-ride-names/internal/service"
)

type WebhookEvent struct {
	ObjectType string                 `json:"object_type"`
	ObjectID   int64                  `json:"object_id"`
	AspectType string                 `json:"aspect_type"`
	OwnerID    int64                  `json:"owner_id"`
	Updates    map[string]interface{} `json:"updates"`
}

type WebhookHandler struct {
	activityService *service.ActivityService
	verifyToken     string
}

func NewWebhookHandler(activityService *service.ActivityService) *WebhookHandler {
	token := make([]byte, 32)
	rand.Read(token)
	return &WebhookHandler{
		activityService: activityService,
		verifyToken:     hex.EncodeToString(token),
	}
}

func (h *WebhookHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/webhook", h.handleWebhook)
}

func (h *WebhookHandler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Handle webhook validation
		mode := r.URL.Query().Get("hub.mode")
		token := r.URL.Query().Get("hub.verify_token")
		challenge := r.URL.Query().Get("hub.challenge")

		if mode != "subscribe" || token != h.verifyToken {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"hub.challenge": challenge,
		})

	case "POST":
		// Handle webhook events
		var event WebhookEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Process only new activities
		if event.ObjectType == "activity" && event.AspectType == "create" {
			go h.activityService.ProcessNewActivity(event.ObjectID)
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
