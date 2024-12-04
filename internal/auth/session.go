package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"cloud.google.com/go/storage"
)

type SessionStore struct {
	sync.RWMutex
	tokenDir   string
	useGCS     bool
	bucketName string
	gcsClient  *storage.Client
}

func NewSessionStore() *SessionStore {
	store := &SessionStore{}

	// Check if we're running in Google Cloud
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		// Initialize Google Cloud Storage client
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Printf("Warning: Failed to initialize Google Cloud Storage: %v", err)
		} else {
			store.gcsClient = client
			store.useGCS = true
			store.bucketName = os.Getenv("TOKEN_BUCKET_NAME")
			if store.bucketName == "" {
				store.bucketName = fmt.Sprintf("%s-tokens", projectID)
			}
			log.Printf("Using Google Cloud Storage bucket: %s", store.bucketName)
			return store
		}
	}

	// Fallback to local file storage
	tokenDir := os.Getenv("TOKEN_STORE_DIR")
	if tokenDir == "" {
		tokenDir = "tokens"
	}

	if err := os.MkdirAll(tokenDir, 0755); err != nil {
		log.Printf("Error creating tokens directory: %v", err)
		tokenDir = filepath.Join(os.TempDir(), "zoatleta-tokens")
		if err := os.MkdirAll(tokenDir, 0755); err != nil {
			log.Printf("Critical: Failed to create fallback token directory: %v", err)
		}
	}

	store.tokenDir = tokenDir
	log.Printf("Using local token directory: %s", tokenDir)
	return store
}

func (s *SessionStore) SetTokens(userID string, tokens *TokenResponse) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty")
	}
	if tokens == nil {
		return fmt.Errorf("tokens cannot be nil")
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %v", err)
	}

	if s.useGCS {
		ctx := context.Background()
		bucket := s.gcsClient.Bucket(s.bucketName)
		obj := bucket.Object(fmt.Sprintf("%s.json", userID))

		writer := obj.NewWriter(ctx)
		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to write to GCS: %v", err)
		}
		if err := writer.Close(); err != nil {
			return fmt.Errorf("failed to close GCS writer: %v", err)
		}

		log.Printf("Successfully stored tokens for user %s in GCS", userID)
		return nil
	}

	// Local file storage logic
	s.Lock()
	defer s.Unlock()

	filePath := filepath.Join(s.tokenDir, fmt.Sprintf("%s.json", userID))
	tempFile := filePath + ".tmp"

	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temporary tokens file: %v", err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to save tokens file: %v", err)
	}

	log.Printf("Successfully stored tokens for user %s locally", userID)
	return nil
}

func (s *SessionStore) GetTokens(userID string, config *OAuth2Config) (*TokenResponse, bool) {
	if userID == "" {
		log.Printf("Warning: Attempted to get tokens with empty userID")
		return nil, false
	}

	var data []byte
	var err error

	if s.useGCS {
		ctx := context.Background()
		bucket := s.gcsClient.Bucket(s.bucketName)
		obj := bucket.Object(fmt.Sprintf("%s.json", userID))

		reader, err := obj.NewReader(ctx)
		if err != nil {
			if err == storage.ErrObjectNotExist {
				return nil, false
			}
			log.Printf("Error reading from GCS: %v", err)
			return nil, false
		}
		defer reader.Close()

		data, err = io.ReadAll(reader)
		if err != nil {
			log.Printf("Error reading GCS data: %v", err)
			return nil, false
		}
	} else {
		s.RLock()
		defer s.RUnlock()

		filePath := filepath.Join(s.tokenDir, fmt.Sprintf("%s.json", userID))
		data, err = os.ReadFile(filePath)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Error reading tokens file: %v", err)
			}
			return nil, false
		}
	}

	var tokens TokenResponse
	if err := json.Unmarshal(data, &tokens); err != nil {
		log.Printf("Error unmarshaling tokens: %v", err)
		return nil, false
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		log.Printf("Invalid token data for user %s: access_token or refresh_token is empty", userID)
		return nil, false
	}

	if tokens.IsExpired() {
		log.Printf("Token expired for user %s, attempting refresh", userID)
		if err := tokens.Refresh(config); err != nil {
			log.Printf("Error refreshing token: %v", err)
			return nil, false
		}

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

	if s.useGCS {
		ctx := context.Background()
		bucket := s.gcsClient.Bucket(s.bucketName)
		obj := bucket.Object(fmt.Sprintf("%s.json", key))

		if err := obj.Delete(ctx); err != nil {
			if err != storage.ErrObjectNotExist {
				return fmt.Errorf("failed to delete from GCS: %v", err)
			}
		}
		return nil
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
