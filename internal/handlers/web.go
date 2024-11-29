package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"log"

	"github.com/guisithos/go-ride-names/internal/auth"
	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/service"
	"github.com/guisithos/go-ride-names/internal/strava"
)

type WebHandler struct {
	sessions     *auth.SessionStore
	oauthConfig  *auth.OAuth2Config
	stravaConfig *config.Config
	verifyToken  string
}

func NewWebHandler(sessions *auth.SessionStore, oauthConfig *auth.OAuth2Config, stravaConfig *config.Config) *WebHandler {
	verifyToken := os.Getenv("WEBHOOK_VERIFY_TOKEN")
	if verifyToken == "" {
		log.Println("Warning: WEBHOOK_VERIFY_TOKEN not set")
		// Generate a random token as fallback
		token := make([]byte, 32)
		rand.Read(token)
		verifyToken = hex.EncodeToString(token)
	}

	return &WebHandler{
		sessions:     sessions,
		oauthConfig:  oauthConfig,
		stravaConfig: stravaConfig,
		verifyToken:  verifyToken,
	}
}

func (h *WebHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.handleHome)
	mux.HandleFunc("/dashboard", h.handleDashboard)
	mux.HandleFunc("/rename-activities", h.handleRenameActivities)
	mux.HandleFunc("/subscribe", h.handleSubscribe)
	mux.HandleFunc("/subscription-status", h.handleSubscriptionStatus)
	mux.HandleFunc("/unsubscribe", h.handleUnsubscribe)
}

func (h *WebHandler) validateToken(tokens *auth.TokenResponse) bool {
	if tokens == nil || tokens.AccessToken == "" {
		return false
	}

	// Check if token is expired
	if time.Now().Unix() >= tokens.ExpiresAt {
		// Try to refresh the token
		client := h.createStravaClient(tokens)
		if err := client.RefreshToken(); err != nil {
			log.Printf("Failed to refresh token: %v", err)
			return false
		}
	}

	return true
}

func (h *WebHandler) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Check if user is already authenticated
	if tokens, exists := h.sessions.GetTokens("user"); exists {
		if h.validateToken(tokens) {
			http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
			return
		}
	}

	tmpl, err := template.ParseFiles(filepath.Join("templates", "home.html"))
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func (h *WebHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	tokens, exists := h.sessions.GetTokens("user")
	if !exists || !h.validateToken(tokens) {
		log.Printf("Invalid or expired tokens, redirecting to home")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join("templates", "dashboard.html"))
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		AccessToken string
		ClientID    string
	}{
		AccessToken: tokens.AccessToken,
		ClientID:    h.oauthConfig.ClientID,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func (h *WebHandler) handleRenameActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokens, exists := h.sessions.GetTokens("user")
	if !exists || !h.validateToken(tokens) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	client := h.createStravaClient(tokens)
	activityService := service.NewActivityService(client)

	_, err := activityService.ListActivities(1, 30, 0, 0, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating activities: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebHandler) createStravaClient(tokens *auth.TokenResponse) *strava.Client {
	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)

	// Set up token refresh callback
	client.SetTokenRefreshCallback(func(newTokens strava.TokenResponse) error {
		return h.sessions.SetTokens("user", &auth.TokenResponse{
			TokenType:    newTokens.TokenType,
			AccessToken:  newTokens.AccessToken,
			RefreshToken: newTokens.RefreshToken,
			ExpiresAt:    newTokens.ExpiresAt,
		})
	})

	return client
}

func (h *WebHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokens, exists := h.sessions.GetTokens("user")
	if !exists || !h.validateToken(tokens) {
		log.Printf("No valid tokens found in session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get base URL from request or environment
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://" + r.Host
	}
	callbackURL := baseURL + "/webhook"
	log.Printf("Subscription attempt - Base URL: %s, Callback URL: %s", baseURL, callbackURL)

	client := h.createStravaClient(tokens)
	activityService := service.NewActivityService(client)

	err := activityService.SubscribeToWebhooks(callbackURL, h.verifyToken)
	if err != nil {
		log.Printf("Error managing webhook subscription: %v", err)
		http.Error(w, fmt.Sprintf("Error managing subscription: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Webhook subscription is active")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": true})
}

func (h *WebHandler) handleSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokens, exists := h.sessions.GetTokens("user")
	if !exists || !h.validateToken(tokens) {
		log.Printf("No valid tokens found in session")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active": false,
			"error":  "No valid authentication tokens found",
		})
		return
	}

	client := h.createStravaClient(tokens)
	activityService := service.NewActivityService(client)

	active, lastCheck := activityService.GetWebhookStatus()
	response := map[string]interface{}{
		"active":    active,
		"lastCheck": lastCheck,
		"timestamp": time.Now().Unix(),
	}

	if !active {
		response["error"] = "Webhook subscription is not active"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *WebHandler) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tokens, exists := h.sessions.GetTokens("user")
	if !exists || !h.validateToken(tokens) {
		log.Printf("No valid tokens found in session")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	client := h.createStravaClient(tokens)
	activityService := service.NewActivityService(client)

	if err := activityService.UnsubscribeFromWebhooks(); err != nil {
		log.Printf("Error unsubscribing from webhooks: %v", err)
		http.Error(w, "Failed to unsubscribe", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
