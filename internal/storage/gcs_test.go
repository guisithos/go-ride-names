package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

func TestGCSConnection(t *testing.T) {
	// Get the absolute path to the project root
	projectRoot, err := filepath.Abs("../../")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	// Load .env file
	if err := godotenv.Load(filepath.Join(projectRoot, ".env")); err != nil {
		t.Logf("Warning: .env file not found: %v", err)
	}

	ctx := context.Background()
	bucketName := getEnvOrFallback("GCS_BUCKET_NAME_TEST", os.Getenv("GCS_BUCKET_NAME"))

	// Use absolute path for credentials file
	credentialsFile := filepath.Join(projectRoot,
		getEnvOrFallback("GOOGLE_APPLICATION_CREDENTIALS_TEST",
			os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))

	t.Logf("Project root: %s", projectRoot)
	t.Logf("Using bucket: %s", bucketName)
	t.Logf("Using credentials: %s", credentialsFile)

	// Verify the credentials file exists
	if _, err := os.Stat(credentialsFile); os.IsNotExist(err) {
		t.Fatalf("Credentials file not found at: %s", credentialsFile)
	}

	store, err := NewGCSStore(ctx, bucketName, credentialsFile)
	if err != nil {
		t.Fatalf("Failed to create GCS store: %v", err)
	}
	defer store.Close()

	// Test basic operations
	testKey := "test-connection"
	testValue := "test-value"

	// Test Set
	if err := store.Set(testKey, testValue); err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Test Get
	value, exists := store.Get(testKey)
	if !exists {
		t.Fatal("Value should exist but doesn't")
	}
	if value != testValue {
		t.Fatalf("Expected value %s, got %s", testValue, value)
	}

	// Test Delete
	if err := store.Delete(testKey); err != nil {
		t.Fatalf("Failed to delete value: %v", err)
	}
}

// getEnvOrFallback returns the value of the test environment variable if it exists,
// otherwise falls back to the production variable
func getEnvOrFallback(testKey, prodKey string) string {
	if value := os.Getenv(testKey); value != "" {
		return value
	}
	return prodKey
}
