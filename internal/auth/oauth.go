package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/guisithos/go-ride-names/internal/config"
)

const (
	AuthURL  = "https://www.strava.com/oauth/authorize"
	TokenURL = "https://www.strava.com/oauth/token"
)

type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type TokenResponse struct {
	TokenType    string  `json:"token_type"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresAt    int64   `json:"expires_at"`
	Athlete      Athlete `json:"athlete"`
}

type Athlete struct {
	ID int64 `json:"id"`
}

func (t *TokenResponse) IsExpired() bool {
	return time.Now().Unix() >= t.ExpiresAt
}

func (t *TokenResponse) Refresh(config *OAuth2Config) error {
	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", t.RefreshToken)

	resp, err := http.PostForm(TokenURL, data)
	if err != nil {
		return fmt.Errorf("error refreshing token: %v", err)
	}
	defer resp.Body.Close()

	var newToken TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&newToken); err != nil {
		return fmt.Errorf("error decoding refresh response: %v", err)
	}

	t.AccessToken = newToken.AccessToken
	t.ExpiresAt = newToken.ExpiresAt
	if newToken.RefreshToken != "" {
		t.RefreshToken = newToken.RefreshToken
	}

	return nil
}

func (t *TokenResponse) GetAthleteID() int64 {
	return t.Athlete.ID
}

type OAuthHandler struct {
	config   *OAuth2Config
	sessions *SessionStore
}

func NewOAuthHandler(cfg *config.Config, sessions *SessionStore) *OAuthHandler {
	return &OAuthHandler{
		config: &OAuth2Config{
			ClientID:     cfg.StravaClientID,
			ClientSecret: cfg.StravaClientSecret,
			RedirectURI:  cfg.OAuth.RedirectURI,
		},
		sessions: sessions,
	}
}

func (h *OAuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/auth", h.handleAuth)
	mux.HandleFunc("/callback", h.handleCallback)
}

func (h *OAuthHandler) handleAuth(w http.ResponseWriter, r *http.Request) {
	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=read,read_all,profile:read_all,activity:read_all,activity:write&approval_prompt=force",
		AuthURL,
		h.config.ClientID,
		url.QueryEscape(h.config.RedirectURI))

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== START OAuth Callback ===")
	log.Printf("Full callback URL: %s", r.URL.String())

	// Check for error parameter from Strava
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		log.Printf("Error from Strava: %s", errMsg)
		http.Error(w, fmt.Sprintf("Authorization failed: %s", errMsg), http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("No authorization code in callback")
		http.Error(w, "Authorization code not found in callback", http.StatusBadRequest)
		return
	}
	log.Printf("Received authorization code: %s", code)

	// Log the configuration being used
	log.Printf("OAuth Configuration - ClientID: %s, RedirectURI: %s", h.config.ClientID, h.config.RedirectURI)

	tokenResp, err := exchangeCodeForToken(code, h.config)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		http.Error(w, "Failed to complete authentication. Please try again.", http.StatusInternalServerError)
		return
	}

	athleteID := tokenResp.GetAthleteID()
	log.Printf("Got athlete ID: %d", athleteID)

	if athleteID == 0 {
		log.Printf("No athlete ID in token response. Full response: %+v", tokenResp)
		http.Error(w, "Invalid response from Strava. Please try again.", http.StatusInternalServerError)
		return
	}

	// Use athlete ID as the session ID
	sessionID := fmt.Sprintf("%d", athleteID)
	log.Printf("Using athlete ID as session ID: %s", sessionID)

	if err := h.sessions.SetTokens(sessionID, tokenResp); err != nil {
		log.Printf("Failed to store tokens: %v", err)
		http.Error(w, "Failed to store authentication. Please try again.", http.StatusInternalServerError)
		return
	}

	// Set cookie with athlete ID
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 24 * 60 * 60, // 60 days
	})

	log.Printf("=== END OAuth Callback - Success ===")
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func exchangeCodeForToken(code string, config *OAuth2Config) (*TokenResponse, error) {
	log.Printf("=== START Token Exchange ===")
	log.Printf("Using code: %s", code)
	log.Printf("Config - ClientID: %s, RedirectURI: %s", config.ClientID, config.RedirectURI)

	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", config.RedirectURI)

	log.Printf("Making request to Strava token endpoint: %s", TokenURL)
	log.Printf("Request data: %s", data.Encode())

	resp, err := http.PostForm(TokenURL, data)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to make request to Strava: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Strava response status: %d", resp.StatusCode)

	// Read the raw response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	log.Printf("Raw response from Strava: %s", string(body))

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		log.Printf("Strava returned non-200 status: %d", resp.StatusCode)
		return nil, fmt.Errorf("strava returned error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Printf("Failed to parse token response: %v", err)
		return nil, fmt.Errorf("error parsing response: %v, body: %s", err, string(body))
	}

	// Validate token response
	if tokenResp.AccessToken == "" {
		log.Printf("Empty access token in response")
		return nil, fmt.Errorf("received empty access token from Strava")
	}

	if tokenResp.Athlete.ID == 0 {
		log.Printf("No athlete ID in response")
		return nil, fmt.Errorf("no athlete ID in token response")
	}

	log.Printf("=== END Token Exchange - Success ===")
	log.Printf("Token type: %s, Athlete ID: %d", tokenResp.TokenType, tokenResp.Athlete.ID)

	return &tokenResp, nil
}

func (h *OAuthHandler) GetConfig() *OAuth2Config {
	return h.config
}
