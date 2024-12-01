package auth

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type SessionStore struct {
	sync.RWMutex
	values       map[string]interface{}
	redis        *redis.Client
	tokenManager *TokenManager
}

func NewSessionStore(redisURL string) *SessionStore {
	store := &SessionStore{
		values:       make(map[string]interface{}),
		tokenManager: NewTokenManager(redisURL),
	}

	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("Warning: Redis URL invalid, falling back to memory store: %v", err)
			return store
		}

		store.redis = redis.NewClient(opt)
		// Test the connection
		if err := store.redis.Ping(context.Background()).Err(); err != nil {
			log.Printf("Warning: Redis connection failed, falling back to memory store: %v", err)
			store.redis = nil
		} else {
			// Start token expiration checker if Redis is available
			store.tokenManager.StartExpirationChecker(5 * time.Minute)
		}
	}

	return store
}

func (s *SessionStore) SetTokens(userID string, tokens *TokenResponse) error {
	return s.tokenManager.StoreTokens(userID, tokens)
}

func (s *SessionStore) GetTokens(userID string) (*TokenResponse, bool) {
	return s.tokenManager.GetTokens(userID)
}

func (s *SessionStore) Set(key string, value interface{}) error {
	s.Lock()
	defer s.Unlock()

	// Store in Redis if available
	if s.redis != nil {
		log.Printf("Attempting to store in Redis: %s", key)
		data, err := json.Marshal(value)
		if err != nil {
			log.Printf("Error marshaling data for Redis: %v", err)
			return err
		}
		if err := s.redis.Set(context.Background(), key, data, 60*24*time.Hour).Err(); err != nil {
			log.Printf("Error storing in Redis: %v", err)
			return err
		}
		log.Printf("Successfully stored data in Redis for key: %s", key)
		return nil
	}

	// Store in memory
	log.Printf("Storing in memory: %s", key)
	s.values[key] = value
	return nil
}

func (s *SessionStore) Get(key string) interface{} {
	s.RLock()
	defer s.RUnlock()

	// Try Redis first
	if s.redis != nil {
		log.Printf("Attempting to retrieve key from Redis: %s", key)
		data, err := s.redis.Get(context.Background(), key).Bytes()
		if err != nil {
			if err != redis.Nil {
				log.Printf("Error retrieving from Redis: %v", err)
			}
		} else {
			var value interface{}
			if err := json.Unmarshal(data, &value); err == nil {
				log.Printf("Successfully retrieved and unmarshaled data from Redis for key: %s", key)
				return value
			}
			log.Printf("Error unmarshaling Redis data: %v", err)
		}
	}

	// Fallback to memory
	log.Printf("Falling back to memory store for key: %s", key)
	return s.values[key]
}

func (s *SessionStore) Delete(key string) error {
	s.Lock()
	defer s.Unlock()

	// Delete from Redis if available
	if s.redis != nil {
		if err := s.redis.Del(context.Background(), key).Err(); err != nil {
			return err
		}
	}

	// Delete from memory
	delete(s.values, key)
	return nil
}

func (s *SessionStore) Clear(userID string) error {
	if err := s.tokenManager.DeleteTokens(userID); err != nil {
		return err
	}
	return s.Delete("user:" + userID)
}
