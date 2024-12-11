package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

type SessionStore struct {
	sync.RWMutex
	redis *redis.Client
}

func NewSessionStore(redisURL string) *SessionStore {
	store := &SessionStore{}

	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("Warning: Redis URL invalid: %v", err)
			return store
		}

		store.redis = redis.NewClient(opt)
		// Test the connection
		if err := store.redis.Ping(context.Background()).Err(); err != nil {
			log.Printf("Warning: Redis connection failed: %v", err)
			store.redis = nil
		} else {
			log.Printf("Successfully connected to Redis")
		}
	}

	return store
}

func (s *SessionStore) SetTokens(athleteID string, tokens *TokenResponse) error {
	if athleteID == "" {
		return fmt.Errorf("athlete ID cannot be empty")
	}
	if tokens == nil {
		return fmt.Errorf("tokens cannot be nil")
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %v", err)
	}

	key := fmt.Sprintf("athlete:%s:tokens", athleteID)
	ctx := context.Background()

	if s.redis != nil {
		if err := s.redis.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
			return fmt.Errorf("failed to store tokens in Redis: %v", err)
		}
		log.Printf("Stored tokens for athlete %s in Redis", athleteID)
		return nil
	}

	return fmt.Errorf("no storage backend available")
}

func (s *SessionStore) GetTokens(athleteID string) (*TokenResponse, bool) {
	if athleteID == "" {
		log.Printf("Warning: Attempted to get tokens with empty athlete ID")
		return nil, false
	}

	key := fmt.Sprintf("athlete:%s:tokens", athleteID)
	ctx := context.Background()

	if s.redis != nil {
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			if err != redis.Nil {
				log.Printf("Error retrieving tokens from Redis: %v", err)
			}
			return nil, false
		}

		var tokens TokenResponse
		if err := json.Unmarshal(data, &tokens); err != nil {
			log.Printf("Error unmarshaling tokens: %v", err)
			return nil, false
		}

		// Extend token expiration
		s.redis.Expire(ctx, key, 24*time.Hour)
		return &tokens, true
	}

	return nil, false
}

func (s *SessionStore) DeleteTokens(athleteID string) error {
	if athleteID == "" {
		return fmt.Errorf("athlete ID cannot be empty")
	}

	key := fmt.Sprintf("athlete:%s:tokens", athleteID)
	ctx := context.Background()

	if s.redis != nil {
		if err := s.redis.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to delete tokens: %v", err)
		}
		log.Printf("Deleted tokens for athlete %s", athleteID)
		return nil
	}

	return fmt.Errorf("no storage backend available")
}

// Set stores a value in Redis with expiration
func (s *SessionStore) Set(key string, value interface{}) error {
	if s.redis == nil {
		return fmt.Errorf("no storage backend available")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}

	ctx := context.Background()
	if err := s.redis.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store value in Redis: %v", err)
	}

	return nil
}

// Get retrieves a value from Redis
func (s *SessionStore) Get(key string) (interface{}, bool) {
	if s.redis == nil {
		return nil, false
	}

	ctx := context.Background()
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			log.Printf("Error retrieving value from Redis: %v", err)
		}
		return nil, false
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		log.Printf("Error unmarshaling value: %v", err)
		return nil, false
	}

	return value, true
}
