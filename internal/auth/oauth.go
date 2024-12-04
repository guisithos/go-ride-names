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
	log.Printf("[INFO] Received callback from Strava: %s", r.URL.String())

	// Check for error parameter from Strava
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		log.Printf("[ERROR] Strava returned error in callback: %s", errMsg)
		http.Error(w, fmt.Sprintf("Authorization failed: %s", errMsg), http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("[ERROR] No authorization code in callback URL")
		http.Error(w, "Authorization code not found in callback", http.StatusBadRequest)
		return
	}

	log.Printf("[DEBUG] Callback received with code: %s", code)
	log.Printf("[DEBUG] Using configuration - ClientID: %s, RedirectURI: %s", h.config.ClientID, h.config.RedirectURI)

	tokenResp, err := exchangeCodeForToken(code, h.config)
	if err != nil {
		log.Printf("[ERROR] Token exchange failed: %v", err)
		http.Error(w, "Failed to complete authentication. Please try again.", http.StatusInternalServerError)
		return
	}

	athleteID := tokenResp.GetAthleteID()
	log.Printf("[DEBUG] Retrieved athlete ID: %d", athleteID)

	if athleteID == 0 {
		log.Printf("[ERROR] No athlete ID in token response. Full response: %+v", tokenResp)
		http.Error(w, "Invalid response from Strava. Please try again.", http.StatusInternalServerError)
		return
	}

	// Use athlete ID as the session ID
	sessionID := fmt.Sprintf("%d", athleteID)
	log.Printf("[DEBUG] Using athlete ID as session ID: %s", sessionID)

	if err := h.sessions.SetTokens(sessionID, tokenResp); err != nil {
		log.Printf("[ERROR] Failed to store tokens: %v", err)
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

	log.Printf("[INFO] Successfully completed OAuth flow for athlete %d", athleteID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func exchangeCodeForToken(code string, config *OAuth2Config) (*TokenResponse, error) {
	log.Printf("[DEBUG] Starting token exchange with code: %s", code)
	log.Printf("[DEBUG] Using client_id: %s and redirect_uri: %s", config.ClientID, config.RedirectURI)

	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", config.RedirectURI)

	log.Printf("[DEBUG] Making POST request to %s with data: %s", TokenURL, data.Encode())

	resp, err := http.PostForm(TokenURL, data)
	if err != nil {
		log.Printf("[ERROR] Failed to make request to Strava: %v", err)
		return nil, fmt.Errorf("failed to make request to Strava: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("[DEBUG] Strava response status: %d", resp.StatusCode)

	// Read the raw response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read response body: %v", err)
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	log.Printf("[DEBUG] Raw response from Strava: %s", string(body))

	// Check for error response
	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] Strava returned error status %d with body: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("strava returned error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Printf("[ERROR] Failed to parse JSON response: %v. Raw response: %s", err, string(body))
		return nil, fmt.Errorf("error parsing response: %v, body: %s", err, string(body))
	}

	// Log the entire token response for debugging
	log.Printf("[DEBUG] Parsed token response: %+v", tokenResp)

	// Validate token response
	if tokenResp.AccessToken == "" {
		log.Printf("[ERROR] Empty access token in response. Full response: %+v", tokenResp)
		return nil, fmt.Errorf("received empty access token from Strava")
	}

	if tokenResp.Athlete.ID == 0 {
		log.Printf("[ERROR] No athlete ID in token response. Full response: %+v", tokenResp)
		return nil, fmt.Errorf("no athlete ID in token response")
	}

	log.Printf("[INFO] Successfully exchanged code for token. Athlete ID: %d", tokenResp.Athlete.ID)
	return &tokenResp, nil
}

func (h *OAuthHandler) GetConfig() *OAuth2Config {
	return h.config
}
