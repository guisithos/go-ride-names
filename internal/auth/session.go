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
		log.Printf("Warning: Could not create tokens directory: %v", err)
	}

	return &SessionStore{
		tokenDir: tokenDir,
	}
}

func (s *SessionStore) SetTokens(userID string, tokens *TokenResponse) error {
	s.Lock()
	defer s.Unlock()

	// Create file path
	filePath := filepath.Join(s.tokenDir, fmt.Sprintf("%s.json", userID))

	// Marshal tokens to JSON
	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write tokens file: %v", err)
	}

	log.Printf("Successfully stored tokens for user %s", userID)
	return nil
}

func (s *SessionStore) GetTokens(userID string, config *OAuth2Config) (*TokenResponse, bool) {
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
	s.Lock()
	defer s.Unlock()

	filePath := filepath.Join(s.tokenDir, fmt.Sprintf("%s.json", key))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *SessionStore) Clear(userID string) error {
	return s.Delete(userID)
}
