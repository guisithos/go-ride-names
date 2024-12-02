package auth

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
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
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf(`{"error": "No code received from Strava"}`)
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	log.Printf(`{"message": "Received callback from Strava", "code": "%s"}`, code)

	tokenResp, err := exchangeCodeForToken(code, h.config)
	if err != nil {
		log.Printf(`{"error": "Failed to exchange code for token", "details": "%v"}`, err)
		http.Error(w, fmt.Sprintf("Error exchanging code: %v", err), http.StatusInternalServerError)
		return
	}

	athleteID := tokenResp.GetAthleteID()
	log.Printf(`{"message": "Got athlete ID", "id": %d}`, athleteID)

	if athleteID == 0 {
		log.Printf(`{"error": "No athlete ID in token response", "response": %+v}`, tokenResp)
		http.Error(w, "Invalid token response", http.StatusInternalServerError)
		return
	}

	// Use athlete ID as the session ID
	sessionID := fmt.Sprintf("%d", athleteID)
	log.Printf(`{"message": "Using athlete ID as session ID", "session_id": "%s"}`, sessionID)

	if err := h.sessions.SetTokens(sessionID, tokenResp); err != nil {
		log.Printf(`{"error": "Failed to store tokens", "details": "%v"}`, err)
		http.Error(w, "Error storing session", http.StatusInternalServerError)
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

	log.Printf(`{"message": "Successfully completed callback handling", "session_id": "%s"}`, sessionID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getDomain(r *http.Request) string {
	host := r.Host
	return strings.TrimPrefix(host, "www.")
}

func exchangeCodeForToken(code string, config *OAuth2Config) (*TokenResponse, error) {
	log.Printf(`{"message": "Starting token exchange", "code": "%s", "client_id": "%s"}`, code, config.ClientID)

	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(TokenURL, data)
	if err != nil {
		log.Printf(`{"error": "Failed to make request to Strava", "details": "%v"}`, err)
		return nil, fmt.Errorf("failed to make request to Strava: %v", err)
	}
	defer resp.Body.Close()

	log.Printf(`{"message": "Got response from Strava", "status": %d}`, resp.StatusCode)

	// Read the raw response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf(`{"error": "Failed to read response body", "details": "%v"}`, err)
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	log.Printf(`{"message": "Raw response from Strava", "response": %s}`, string(body))

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		log.Printf(`{"error": "Failed to parse token response", "details": "%v", "body": %s}`, err, string(body))
		return nil, fmt.Errorf("error parsing response: %v, body: %s", err, string(body))
	}

	log.Printf(`{"message": "Successfully parsed token response", "token_type": "%s", "athlete": %+v}`,
		tokenResp.TokenType, tokenResp.Athlete)

	return &tokenResp, nil
}

func (h *OAuthHandler) GetConfig() *OAuth2Config {
	return h.config
}
