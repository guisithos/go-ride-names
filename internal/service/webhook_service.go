package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/guisithos/go-ride-names/internal/strava"
)

type WebhookService struct {
	client              *strava.Client
	webhookSubscription *strava.WebhookSubscription
}

func NewWebhookService(client *strava.Client) *WebhookService {
	return &WebhookService{
		client: client,
	}
}

func (s *WebhookService) SubscribeToWebhooks(callbackURL, verifyToken string) error {
	log.Printf("Starting webhook subscription process for URL: %s", callbackURL)

	// First, list and delete ALL existing subscriptions
	subscriptions, err := s.client.ListWebhookSubscriptions()
	if err != nil {
		return fmt.Errorf("error listing subscriptions: %v", err)
	}

	// Delete ALL existing subscriptions
	for _, sub := range subscriptions {
		log.Printf("Deleting subscription ID: %d with URL: %s", sub.ID, sub.CallbackURL)
		if err := s.client.DeleteWebhookSubscription(sub.ID); err != nil {
			// Log but continue if deletion fails
			log.Printf("Warning: failed to delete subscription %d: %v", sub.ID, err)
		}
		// Add a small delay between deletions
		time.Sleep(100 * time.Millisecond)
	}

	// Wait a moment after deletions
	time.Sleep(500 * time.Millisecond)

	// Create new subscription
	log.Printf("Creating new webhook subscription with URL: %s", callbackURL)
	subscription, err := s.client.CreateWebhookSubscription(callbackURL, verifyToken)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			// If it already exists, try to delete all subscriptions again
			log.Printf("Subscription exists, retrying cleanup...")
			time.Sleep(1 * time.Second)
			return s.SubscribeToWebhooks(callbackURL, verifyToken)
		}
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	log.Printf("Successfully created new webhook subscription: ID=%d with URL: %s",
		subscription.ID, callbackURL)
	s.webhookSubscription = subscription
	return nil
}

func (s *WebhookService) GetSubscriptionStatus() (bool, error) {
	subscriptions, err := s.client.ListWebhookSubscriptions()
	if err != nil {
		return false, err
	}

	// Log subscription details
	for _, sub := range subscriptions {
		log.Printf("Found subscription: ID=%d, CallbackURL=%s",
			sub.ID, sub.CallbackURL)
	}

	return len(subscriptions) > 0, nil
}

func (s *WebhookService) UnsubscribeFromWebhooks() error {
	subscriptions, err := s.client.ListWebhookSubscriptions()
	if err != nil {
		return fmt.Errorf("error listing subscriptions: %v", err)
	}

	for _, sub := range subscriptions {
		log.Printf("Deleting subscription ID: %d", sub.ID)
		if err := s.client.DeleteWebhookSubscription(sub.ID); err != nil {
			return fmt.Errorf("error deleting subscription %d: %v", sub.ID, err)
		}
	}

	return nil
}
