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

const (
	// Token will expire 1 hour before the actual expiration to ensure we have time to refresh
	tokenExpirationBuffer = 1 * time.Hour
	// Default token storage duration in Redis (60 days)
	defaultTokenStorageDuration = 60 * 24 * time.Hour
)

type TokenManager struct {
	mu            sync.RWMutex
	redis         *redis.Client
	refreshTokens map[string]string // In-memory backup of refresh tokens
}

func NewTokenManager(redisURL string) *TokenManager {
	tm := &TokenManager{
		refreshTokens: make(map[string]string),
	}

	if redisURL != "" {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("Warning: Redis URL invalid, falling back to memory store: %v", err)
			return tm
		}

		tm.redis = redis.NewClient(opt)
		// Test the connection
		if err := tm.redis.Ping(context.Background()).Err(); err != nil {
			log.Printf("Warning: Redis connection failed, falling back to memory store: %v", err)
			tm.redis = nil
		}
	}

	return tm
}

func (tm *TokenManager) StoreTokens(userID string, tokens *TokenResponse) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Store refresh token in memory as backup
	tm.refreshTokens[userID] = tokens.RefreshToken

	// Calculate token expiration
	expiresIn := time.Until(time.Unix(tokens.ExpiresAt, 0))
	if expiresIn < 0 {
		expiresIn = 1 * time.Hour // Default to 1 hour if token is already expired
	}

	// Store in Redis with longer expiration
	if tm.redis != nil {
		data, err := json.Marshal(tokens)
		if err != nil {
			return fmt.Errorf("failed to marshal tokens: %v", err)
		}

		// Store tokens with a longer expiration to keep refresh tokens available
		if err := tm.redis.Set(context.Background(),
			fmt.Sprintf("tokens:%s", userID),
			data,
			defaultTokenStorageDuration).Err(); err != nil {
			return fmt.Errorf("failed to store tokens in Redis: %v", err)
		}

		// Set expiration reminder
		if err := tm.redis.Set(context.Background(),
			fmt.Sprintf("token_expiration:%s", userID),
			tokens.ExpiresAt,
			expiresIn-tokenExpirationBuffer).Err(); err != nil {
			log.Printf("Warning: Failed to set token expiration reminder: %v", err)
		}
	}

	return nil
}

func (tm *TokenManager) GetTokens(userID string) (*TokenResponse, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	log.Printf("Getting tokens for user ID: %s", userID)

	if tm.redis != nil {
		// Try to get tokens from Redis
		data, err := tm.redis.Get(context.Background(), fmt.Sprintf("tokens:%s", userID)).Bytes()
		if err == nil {
			var tokens TokenResponse
			if err := json.Unmarshal(data, &tokens); err == nil {
				log.Printf("Found tokens in Redis for user %s, expires at %d", userID, tokens.ExpiresAt)
				// Only return false if token is actually expired
				if time.Now().Unix() >= tokens.ExpiresAt {
					log.Printf("Token is expired for user %s", userID)
					return &tokens, false
				}
				return &tokens, true
			}
			log.Printf("Error unmarshaling tokens: %v", err)
		} else {
			log.Printf("Error getting tokens from Redis: %v", err)
		}
	}

	// If we have a refresh token in memory, return it
	if refreshToken, exists := tm.refreshTokens[userID]; exists {
		log.Printf("Found refresh token in memory for user %s", userID)
		return &TokenResponse{
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Unix(), // Expired token to force refresh
		}, false
	}

	log.Printf("No tokens found for user %s", userID)
	return nil, false
}

func (tm *TokenManager) DeleteTokens(userID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Remove from memory
	delete(tm.refreshTokens, userID)

	// Remove from Redis
	if tm.redis != nil {
		ctx := context.Background()
		if err := tm.redis.Del(ctx,
			fmt.Sprintf("tokens:%s", userID),
			fmt.Sprintf("token_expiration:%s", userID)).Err(); err != nil {
			return fmt.Errorf("failed to delete tokens from Redis: %v", err)
		}
	}

	return nil
}

func (tm *TokenManager) StartExpirationChecker(checkInterval time.Duration) {
	if tm.redis == nil {
		return // Only run with Redis
	}

	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for range ticker.C {
			tm.checkExpiredTokens()
		}
	}()
}

func (tm *TokenManager) checkExpiredTokens() {
	ctx := context.Background()
	pattern := "token_expiration:*"

	iter := tm.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		userID := key[len("token_expiration:"):]

		// If we find an expiration key, it means the token needs refresh
		tokens, exists := tm.GetTokens(userID)
		if exists && tokens != nil {
			log.Printf("Token for user %s needs refresh", userID)
			// The next GetTokens call will trigger a refresh
		}
	}
}
