package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func StartOAuthFlow(clientID, clientSecret string) {
	config := &OAuth2Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  "http://localhost:8080/callback",
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=read,read_all,profile:read_all,activity:read_all,activity:write&approval_prompt=force",
			AuthURL,
			config.ClientID,
			url.QueryEscape(config.RedirectURI))

		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}

		// Exchange code for token
		tokenResp, err := exchangeCodeForToken(code, config)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error exchanging code: %v", err), http.StatusInternalServerError)
			return
		}

		// Make the response more readable and copyable
		response := fmt.Sprintf(`
Successfully authenticated! Add these values to your .env file:

STRAVA_CLIENT_ID=%s
STRAVA_CLIENT_SECRET=%s
STRAVA_ACCESS_TOKEN=%s
STRAVA_REFRESH_TOKEN=%s

Token expires at: %d
`,
			config.ClientID,
			config.ClientSecret,
			tokenResp.AccessToken,
			tokenResp.RefreshToken,
			tokenResp.ExpiresAt,
		)

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, response)
	})

	fmt.Println("Starting server on :8080")
	fmt.Println("Please visit http://localhost:8080 to begin OAuth flow")
	http.ListenAndServe(":8080", nil)
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
