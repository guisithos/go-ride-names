package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

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
}

type SessionStore struct {
	sync.RWMutex
	tokens map[string]*TokenResponse
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		tokens: make(map[string]*TokenResponse),
	}
}

func (s *SessionStore) GetTokens(userID string) (*TokenResponse, bool) {
	s.RLock()
	defer s.RUnlock()
	tokens, exists := s.tokens[userID]
	return tokens, exists
}

func (s *SessionStore) SetTokens(userID string, tokens *TokenResponse) {
	s.Lock()
	defer s.Unlock()
	s.tokens[userID] = tokens
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
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	tokenResp, err := exchangeCodeForToken(code, h.config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error exchanging code: %v", err), http.StatusInternalServerError)
		return
	}

	h.sessions.SetTokens("user", tokenResp)
	http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
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
