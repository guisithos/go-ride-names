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
}

func NewWebHandler(store storage.Store, oauthCfg *auth.OAuth2Config, stravaConfig *config.Config) *WebHandler {
	return &WebHandler{
		store:        store,
		oauthCfg:     oauthCfg,
		stravaConfig: stravaConfig,
	}
}

func (h *WebHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.handleHome)
	mux.HandleFunc("/dashboard", h.handleDashboard)
	mux.HandleFunc("/rename-activities", h.handleRenameActivities)
	mux.HandleFunc("/subscribe", h.handleSubscribe)
	mux.HandleFunc("/subscription-status", h.handleSubscriptionStatus)
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
	// Get athlete ID from cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	athleteID := cookie.Value
	tokensInterface, exists := h.store.GetTokens(athleteID)
	if !exists {
		log.Printf("No tokens found for athlete %s", athleteID)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	tokens, ok := tokensInterface.(*auth.TokenResponse)
	if !ok {
		log.Printf("Invalid token type for athlete %s", athleteID)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Check if token needs refresh
	now := time.Now().Unix()
	if now >= tokens.ExpiresAt {
		log.Printf("Token expired for athlete %s, refreshing...", athleteID)
		client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
			h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)

		newTokens, err := client.RefreshToken()
		if err != nil {
			log.Printf("Failed to refresh token: %v", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		// Update tokens in session
		tokens.AccessToken = newTokens.AccessToken
		tokens.RefreshToken = newTokens.RefreshToken
		tokens.ExpiresAt = newTokens.ExpiresAt

		if err := h.store.SetTokens(athleteID, tokens); err != nil {
			log.Printf("Failed to update tokens: %v", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
	}

	tmpl, err := template.ParseFiles(filepath.Join("templates", "dashboard.html"))
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := struct {
		AccessToken string
		AthleteID   string
	}{
		AccessToken: tokens.AccessToken,
		AthleteID:   athleteID,
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

	tokens, ok := tokensInterface.(*auth.TokenResponse)
	if !ok {
		log.Printf("Invalid token type for athlete %s", athleteID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)
	activityService := service.NewActivityService(client)

	_, err = activityService.ListActivities(1, 30, 0, 0, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error updating activities: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	tokens, ok := tokensInterface.(*auth.TokenResponse)
	if !ok {
		log.Printf("Invalid token type for athlete %s", athleteID)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
	activityService := service.NewActivityService(client)

	err = activityService.SubscribeToWebhooks(callbackURL, verifyToken)
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

	tokens, ok := tokensInterface.(*auth.TokenResponse)
	if !ok {
		log.Printf("Invalid token type for athlete %s", athleteID)
		json.NewEncoder(w).Encode(map[string]bool{"active": false})
		return
	}

	client := strava.NewClient(tokens.AccessToken, tokens.RefreshToken,
		h.stravaConfig.StravaClientID, h.stravaConfig.StravaClientSecret)

	// Check actual subscription status
	subs, err := client.ListWebhookSubscriptions()
	if err != nil {
		log.Printf("Error checking subscriptions: %v", err)
		json.NewEncoder(w).Encode(map[string]bool{"active": false})
		return
	}

	// Check if we have any active subscriptions
	active := len(subs) > 0
	if active {
		// Update session status
		if err := h.store.Set(fmt.Sprintf("webhook_active:%s", athleteID), true); err != nil {
			log.Printf("Error storing webhook status: %v", err)
		}
	} else {
		h.store.Set(fmt.Sprintf("webhook_active:%s", athleteID), nil)
	}

	log.Printf("Subscription status check for athlete %s - Active: %v, Subscriptions: %d",
		athleteID, active, len(subs))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"active": active})
}
