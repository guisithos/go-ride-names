package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type GCSStore struct {
	client     *storage.Client
	bucketName string
	ctx        context.Context
}

func NewGCSStore(ctx context.Context, bucketName string, credentialsFile string) (*GCSStore, error) {
	var client *storage.Client
	var err error

	if credentialsFile != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsFile))
	} else {
		// Use default credentials
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	store := &GCSStore{
		client:     client,
		bucketName: bucketName,
		ctx:        ctx,
	}

	// Verify bucket exists and is accessible
	if err := store.verifyBucket(); err != nil {
		client.Close()
		return nil, err
	}

	return store, nil
}

func (s *GCSStore) verifyBucket() error {
	bucket := s.client.Bucket(s.bucketName)
	_, err := bucket.Attrs(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to access bucket %s: %v", s.bucketName, err)
	}
	return nil
}

func (s *GCSStore) Close() error {
	return s.client.Close()
}

func (s *GCSStore) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("ERROR: Failed to marshal value: %v", err)
		return fmt.Errorf("marshal error: %v", err)
	}

	obj := s.client.Bucket(s.bucketName).Object(key)
	w := obj.NewWriter(s.ctx)

	log.Printf("DEBUG: Writing to key: %s", key)
	if _, err := w.Write(data); err != nil {
		log.Printf("ERROR: Failed to write to GCS: %v", err)
		return fmt.Errorf("write error: %v", err)
	}

	if err := w.Close(); err != nil {
		log.Printf("ERROR: Failed to close GCS writer: %v", err)
		return fmt.Errorf("close error: %v", err)
	}

	return nil
}

func (s *GCSStore) Get(key string) (interface{}, bool) {
	obj := s.client.Bucket(s.bucketName).Object(key)
	r, err := obj.NewReader(s.ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, false
		}
		log.Printf("Error reading from GCS: %v", err)
		return nil, false
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		log.Printf("Error reading data: %v", err)
		return nil, false
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		log.Printf("Error unmarshaling value: %v", err)
		return nil, false
	}

	return value, true
}

func (s *GCSStore) Delete(key string) error {
	obj := s.client.Bucket(s.bucketName).Object(key)
	if err := obj.Delete(s.ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return nil
		}
		return fmt.Errorf("failed to delete object: %v", err)
	}
	return nil
}

// TokenStore implementation
func (s *GCSStore) SetTokens(athleteID string, tokens interface{}) error {
	key := fmt.Sprintf("athlete/%s/tokens.json", athleteID)
	log.Printf("DEBUG: Attempting to store tokens for athlete %s", athleteID)
	log.Printf("DEBUG: Using bucket: %s", s.bucketName)

	err := s.Set(key, tokens)
	if err != nil {
		log.Printf("ERROR: Failed to store tokens in GCS: %v", err)
		return fmt.Errorf("storage error: %v", err)
	}

	log.Printf("SUCCESS: Stored tokens for athlete %s", athleteID)
	return nil
}

func (s *GCSStore) GetTokens(athleteID string) (interface{}, bool) {
	if athleteID == "" {
		log.Printf("Warning: Attempted to get tokens with empty athlete ID")
		return nil, false
	}

	key := fmt.Sprintf("athlete/%s/tokens.json", athleteID)
	log.Printf("DEBUG: Retrieving tokens for athlete %s", athleteID)

	value, exists := s.Get(key)
	if !exists {
		log.Printf("DEBUG: No tokens found for athlete %s", athleteID)
		return nil, false
	}

	log.Printf("DEBUG: Successfully retrieved tokens for athlete %s", athleteID)
	return value, true
}

func (s *GCSStore) DeleteTokens(athleteID string) error {
	if athleteID == "" {
		return fmt.Errorf("athlete ID cannot be empty")
	}

	key := fmt.Sprintf("athlete/%s/tokens.json", athleteID)
	return s.Delete(key)
}
