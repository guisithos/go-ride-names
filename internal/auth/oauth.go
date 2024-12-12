package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/guisithos/go-ride-names/internal/config"
	"github.com/guisithos/go-ride-names/internal/storage"
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

type Athlete struct {
	ID int64 `json:"id"`
}

type TokenResponse struct {
	TokenType    string  `json:"token_type"`
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresAt    int64   `json:"expires_at"`
	Athlete      Athlete `json:"athlete"`
}

type OAuthHandler struct {
	config *OAuth2Config
	store  storage.Store
}

func NewOAuthHandler(cfg *config.Config, store storage.Store) *OAuthHandler {
	return &OAuthHandler{
		config: &OAuth2Config{
			ClientID:     cfg.StravaClientID,
			ClientSecret: cfg.StravaClientSecret,
			RedirectURI:  cfg.OAuth.RedirectURI,
		},
		store: store,
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
		log.Printf("No authorization code in callback")
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	tokenResp, err := exchangeCodeForToken(code, h.config)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		http.Error(w, fmt.Sprintf("Error exchanging code: %v", err), http.StatusInternalServerError)
		return
	}

	if tokenResp.Athlete.ID == 0 {
		log.Printf("No athlete ID in token response")
		http.Error(w, "Invalid token response", http.StatusInternalServerError)
		return
	}

	// Store the entire TokenResponse object
	sessionKey := fmt.Sprintf("%d", tokenResp.Athlete.ID)
	if err := h.store.SetTokens(sessionKey, tokenResp); err != nil {
		log.Printf("Failed to store tokens: %v", err)
		http.Error(w, "Failed to store authentication", http.StatusInternalServerError)
		return
	}

	// Set cookie with athlete ID
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionKey,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   60 * 24 * 60 * 60, // 60 days
	})

	log.Printf("Successfully authenticated athlete %d", tokenResp.Athlete.ID)
	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
}

func exchangeCodeForToken(code string, config *OAuth2Config) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", config.RedirectURI)

	resp, err := http.PostForm(TokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Strava: %v", err)
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &tokenResp, nil
}

func (h *OAuthHandler) GetConfig() *OAuth2Config {
	return h.config
}
