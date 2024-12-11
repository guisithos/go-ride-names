package storage

// Store defines the interface for storage implementations
type Store interface {
	// Generic key-value operations
	Set(key string, value interface{}) error
	Get(key string) (interface{}, bool)
	Delete(key string) error

	// Token-specific operations
	SetTokens(athleteID string, tokens interface{}) error
	GetTokens(athleteID string) (interface{}, bool)
	DeleteTokens(athleteID string) error

	// Cleanup
	Close() error
}
