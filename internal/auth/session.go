package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type SessionStore struct {
	sync.RWMutex
	tokenDir string
}

func NewSessionStore(redisURL string) *SessionStore {
	// Create tokens directory if it doesn't exist
	tokenDir := "tokens"
	if err := os.MkdirAll(tokenDir, 0755); err != nil {
		log.Printf("Error creating tokens directory: %v", err)
		// Try to use a temporary directory as fallback
		tokenDir = filepath.Join(os.TempDir(), "zoatleta-tokens")
		if err := os.MkdirAll(tokenDir, 0755); err != nil {
			log.Printf("Critical: Failed to create fallback token directory: %v", err)
		}
	}

	log.Printf("Using token directory: %s", tokenDir)
	return &SessionStore{
		tokenDir: tokenDir,
	}
}

func (s *SessionStore) SetTokens(userID string, tokens *TokenResponse) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}
	if tokens == nil {
		return fmt.Errorf("tokens cannot be nil")
	}

	s.Lock()
	defer s.Unlock()

	// Create file path
	filePath := filepath.Join(s.tokenDir, fmt.Sprintf("%s.json", userID))

	// Marshal tokens to JSON
	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %v", err)
	}

	// Write to temporary file first
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary tokens file: %v", err)
	}

	// Rename temporary file to final file (atomic operation)
	if err := os.Rename(tempFile, filePath); err != nil {
		// Try to clean up the temporary file
		os.Remove(tempFile)
		return fmt.Errorf("failed to save tokens file: %v", err)
	}

	log.Printf("Successfully stored tokens for user %s", userID)
	return nil
}

func (s *SessionStore) GetTokens(userID string, config *OAuth2Config) (*TokenResponse, bool) {
	if userID == "" {
		log.Printf("Warning: Attempted to get tokens with empty userID")
		return nil, false
	}

	s.RLock()
	defer s.RUnlock()

	// Read file
	filePath := filepath.Join(s.tokenDir, fmt.Sprintf("%s.json", userID))
	data, err := os.ReadFile(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error reading tokens file: %v", err)
		}
		return nil, false
	}

	// Unmarshal tokens
	var tokens TokenResponse
	if err := json.Unmarshal(data, &tokens); err != nil {
		log.Printf("Error unmarshaling tokens: %v", err)
		// Try to clean up corrupted file
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to remove corrupted token file: %v", err)
		}
		return nil, false
	}

	// Validate token data
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		log.Printf("Invalid token data for user %s: access_token or refresh_token is empty", userID)
		return nil, false
	}

	// Check if token is expired
	if tokens.IsExpired() {
		log.Printf("Token expired for user %s, attempting refresh", userID)
		if err := tokens.Refresh(config); err != nil {
			log.Printf("Error refreshing token: %v", err)
			return nil, false
		}

		// Save refreshed tokens
		if err := s.SetTokens(userID, &tokens); err != nil {
			log.Printf("Error saving refreshed tokens: %v", err)
			return nil, false
		}
		log.Printf("Successfully refreshed and saved tokens for user %s", userID)
	}

	return &tokens, true
}

func (s *SessionStore) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	s.Lock()
	defer s.Unlock()

	filePath := filepath.Join(s.tokenDir, fmt.Sprintf("%s.json", key))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %v", err)
	}
	return nil
}

func (s *SessionStore) Clear(userID string) error {
	return s.Delete(userID)
}
