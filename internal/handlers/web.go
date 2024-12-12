package handlers

import (
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
	"github.com/guisithos/go-ride-names/internal/storage"
	"github.com/guisithos/go-ride-names/internal/strava"
)

type WebHandler struct {
	store        storage.Store
	oauthCfg     *auth.OAuth2Config
	stravaConfig *config.Config
	templates    *template.Template
}

func NewWebHandler(store storage.Store, oauthCfg *auth.OAuth2Config, stravaConfig *config.Config, templates *template.Template) *WebHandler {
	return &WebHandler{
		store:        store,
		oauthCfg:     oauthCfg,
		stravaConfig: stravaConfig,
		templates:    templates,
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

func (h *WebHandler) handleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
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
	// Get session cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Get tokens from storage
	athleteID := cookie.Value
	tokensInterface, exists := h.store.GetTokens(athleteID)
	if !exists {
		log.Printf("No tokens found for athlete %s", athleteID)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Convert stored tokens to TokenResponse
	tokenData, err := json.Marshal(tokensInterface)
	if err != nil {
		log.Printf("Failed to marshal token data: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var tokens auth.TokenResponse
	if err := json.Unmarshal(tokenData, &tokens); err != nil {
		log.Printf("Failed to unmarshal token data: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Render dashboard template with both athleteID and accessToken
	data := struct {
		AthleteID   string
		AccessToken string
	}{
		AthleteID:   athleteID,
		AccessToken: tokens.AccessToken,
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("Error rendering dashboard template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *WebHandler) handleRenameActivities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get athlete ID from cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	athleteID := cookie.Value
	tokensInterface, exists := h.store.GetTokens(athleteID)
	if !exists {
		log.Printf("No tokens found for athlete %s", athleteID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert the stored token data to TokenResponse
	tokenData, err := json.Marshal(tokensInterface)
	if err != nil {
		log.Printf("Failed to marshal token data: %v", err)
		http.Error(w, "Invalid token data", http.StatusInternalServerError)
		return
	}

	var tokens auth.TokenResponse
	if err := json.Unmarshal(tokenData, &tokens); err != nil {
		log.Printf("Failed to unmarshal token data: %v", err)
		http.Error(w, "Invalid token format", http.StatusInternalServerError)
		return
	}

	// Create Strava client and ActivityService
	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)
	activityService := service.NewActivityService(client)

	// Get recent activities and update their names
	activities, err := activityService.ListActivities(1, 30, 0, 0, true)
	if err != nil {
		log.Printf("Error processing activities: %v", err)
		http.Error(w, "Failed to process activities", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully processed activities for athlete %s", athleteID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"renamed": len(activities),
	})
}

func (h *WebHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Invalid method: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get athlete ID from cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	athleteID := cookie.Value
	tokensInterface, exists := h.store.GetTokens(athleteID)
	if !exists {
		log.Printf("No tokens found for athlete %s", athleteID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert stored tokens to TokenResponse
	tokenData, err := json.Marshal(tokensInterface)
	if err != nil {
		log.Printf("Failed to marshal token data: %v", err)
		http.Error(w, "Invalid token data", http.StatusInternalServerError)
		return
	}

	var tokens auth.TokenResponse
	if err := json.Unmarshal(tokenData, &tokens); err != nil {
		log.Printf("Failed to unmarshal token data: %v", err)
		http.Error(w, "Invalid token format", http.StatusInternalServerError)
		return
	}

	// Check if token needs refresh
	now := time.Now().Unix()
	if now >= tokens.ExpiresAt {
		client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
			h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)

		newTokens, err := client.RefreshToken()
		if err != nil {
			log.Printf("Failed to refresh token: %v", err)
			http.Error(w, "Authentication error", http.StatusUnauthorized)
			return
		}

		tokens.AccessToken = newTokens.AccessToken
		tokens.RefreshToken = newTokens.RefreshToken
		tokens.ExpiresAt = newTokens.ExpiresAt

		if err := h.store.SetTokens(athleteID, tokens); err != nil {
			log.Printf("Failed to update tokens: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
	}

	// Get base URL from request or environment
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://" + r.Host
	}
	callbackURL := baseURL + "/webhook"

	// Validate Strava configuration
	if h.stravaConfig.StravaClientID == "" || h.stravaConfig.StravaClientSecret == "" {
		log.Printf("Error: Strava credentials not configured properly")
		http.Error(w, "Strava configuration error", http.StatusInternalServerError)
		return
	}

	verifyToken := os.Getenv("WEBHOOK_VERIFY_TOKEN")
	if verifyToken == "" {
		log.Printf("Error: WEBHOOK_VERIFY_TOKEN not configured")
		http.Error(w, "Webhook verify token not configured", http.StatusInternalServerError)
		return
	}

	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)
	webhookService := service.NewWebhookService(client)

	err = webhookService.SubscribeToWebhooks(callbackURL, verifyToken)
	if err != nil {
		log.Printf("Error managing webhook subscription: %v", err)
		http.Error(w, fmt.Sprintf("Error managing subscription: %v", err), http.StatusInternalServerError)
		return
	}

	// Store subscription status in session with longer expiration
	if err := h.store.Set(fmt.Sprintf("webhook_active:%s", athleteID), true); err != nil {
		log.Printf("Error storing webhook status: %v", err)
	}

	log.Printf("Webhook subscription is active for athlete %s", athleteID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": true})
}

func (h *WebHandler) handleSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get athlete ID from cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		json.NewEncoder(w).Encode(map[string]bool{"active": false})
		return
	}

	athleteID := cookie.Value
	tokensInterface, exists := h.store.GetTokens(athleteID)
	if !exists {
		log.Printf("No tokens found for athlete %s", athleteID)
		json.NewEncoder(w).Encode(map[string]bool{"active": false})
		return
	}

	// Convert stored tokens to TokenResponse
	tokenData, err := json.Marshal(tokensInterface)
	if err != nil {
		log.Printf("Failed to marshal token data: %v", err)
		json.NewEncoder(w).Encode(map[string]bool{"active": false})
		return
	}

	var tokens auth.TokenResponse
	if err := json.Unmarshal(tokenData, &tokens); err != nil {
		log.Printf("Failed to unmarshal token data: %v", err)
		json.NewEncoder(w).Encode(map[string]bool{"active": false})
		return
	}

	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)

	// Check webhook subscriptions
	subs, err := client.ListWebhookSubscriptions()
	if err != nil {
		log.Printf("Error checking subscriptions: %v", err)
		json.NewEncoder(w).Encode(map[string]bool{"active": false})
		return
	}

	// Consider any subscription as active, regardless of URL
	active := len(subs) > 0

	// Update stored status
	statusKey := fmt.Sprintf("webhook_active:%s", athleteID)
	if err := h.store.Set(statusKey, active); err != nil {
		log.Printf("Error storing webhook status: %v", err)
	}

	log.Printf("Subscription status for athlete %s: Active=%v, Subs=%d",
		athleteID, active, len(subs))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": active})
}

func (h *WebHandler) handleUnsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	athleteID := cookie.Value
	tokensInterface, exists := h.store.GetTokens(athleteID)
	if !exists {
		log.Printf("No tokens found for athlete %s", athleteID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokenData, err := json.Marshal(tokensInterface)
	if err != nil {
		log.Printf("Failed to marshal token data: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	var tokens auth.TokenResponse
	if err := json.Unmarshal(tokenData, &tokens); err != nil {
		log.Printf("Failed to unmarshal token data: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)
	webhookService := service.NewWebhookService(client)

	if err := webhookService.UnsubscribeFromWebhooks(); err != nil {
		log.Printf("Error unsubscribing from webhooks: %v", err)
		http.Error(w, "Failed to unsubscribe", http.StatusInternalServerError)
		return
	}

	// Remove subscription status from store
	statusKey := fmt.Sprintf("webhook_active:%s", athleteID)
	h.store.Set(statusKey, nil)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": false})
}
