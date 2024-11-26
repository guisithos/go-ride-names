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
	tokens map[string]*TokenResponse
	values map[string]interface{}
	redis  *redis.Client
}

func NewSessionStore(redisURL string) *SessionStore {
	store := &SessionStore{
		tokens: make(map[string]*TokenResponse),
		values: make(map[string]interface{}),
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
		}
	}

	return store
}

func (s *SessionStore) SetTokens(userID string, tokens *TokenResponse) error {
	s.Lock()
	defer s.Unlock()

	// Store in memory
	s.tokens[userID] = tokens

	// Store in Redis if available
	if s.redis != nil {
		data, err := json.Marshal(tokens)
		if err != nil {
			return err
		}
		return s.redis.Set(context.Background(), "tokens:"+userID, data, 24*time.Hour).Err()
	}
	return nil
}

func (s *SessionStore) GetTokens(userID string) (*TokenResponse, bool) {
	s.RLock()
	defer s.RUnlock()

	// Try Redis first i
	if s.redis != nil {
		data, err := s.redis.Get(context.Background(), "tokens:"+userID).Bytes()
		if err == nil {
			var tokens TokenResponse
			if err := json.Unmarshal(data, &tokens); err == nil {
				return &tokens, true
			}
		}
	}

	// Fallback to memory
	tokens, exists := s.tokens[userID]
	return tokens, exists
}

func (s *SessionStore) Set(key string, value interface{}) error {
	s.Lock()
	defer s.Unlock()

	// Store in Redis
	if s.redis != nil {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		return s.redis.Set(context.Background(), key, data, 24*time.Hour).Err()
	}

	// Store in memory based on value type
	if tokenResp, ok := value.(*TokenResponse); ok {
		s.tokens[key] = tokenResp
	} else {
		s.values[key] = value
	}
	return nil
}

func (s *SessionStore) Get(key string) interface{} {
	s.RLock()
	defer s.RUnlock()

	// Try Redis first
	if s.redis != nil {
		data, err := s.redis.Get(context.Background(), key).Bytes()
		if err == nil {
			var value interface{}
			if err := json.Unmarshal(data, &value); err == nil {
				return value
			}
			// Try as boolean
			var boolValue bool
			if err := json.Unmarshal(data, &boolValue); err == nil {
				return boolValue
			}
		}
	}

	// Fallback to memory
	// First check tokens map
	if token, exists := s.tokens[key]; exists {
		return token
	}
	// Then check values map
	return s.values[key]
}
