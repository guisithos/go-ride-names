package auth

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

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
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	AthleteID    int64  `json:"athlete.id"`
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
		log.Printf("Error: No code received from Strava")
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	tokenResp, err := exchangeCodeForToken(code, h.config)
	if err != nil {
		log.Printf("Error exchanging code for token: %v", err)
		http.Error(w, fmt.Sprintf("Error exchanging code: %v", err), http.StatusInternalServerError)
		return
	}

	if tokenResp.AthleteID == 0 {
		log.Printf("Error: No athlete ID in token response")
		http.Error(w, "Invalid token response", http.StatusInternalServerError)
		return
	}

	// Use athlete ID as the session ID
	sessionID := fmt.Sprintf("%d", tokenResp.AthleteID)
	log.Printf("Using athlete ID as session ID: %s", sessionID)

	if err := h.sessions.SetTokens(sessionID, tokenResp); err != nil {
		log.Printf("Error storing tokens: %v", err)
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

	log.Printf("Successfully set session cookie and stored tokens, redirecting to dashboard")
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
	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(TokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (h *OAuthHandler) GetConfig() *OAuth2Config {
	return h.config
}
