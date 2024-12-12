package service

import (
	"fmt"
	"log"
	"strings"

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
	log.Printf("Checking existing webhook subscriptions...")

	subscriptions, err := s.client.ListWebhookSubscriptions()
	if err != nil {
		log.Printf("Error listing subscriptions: %v", err)
		return err
	}

	// Delete any existing subscriptions regardless of URL
	for _, sub := range subscriptions {
		log.Printf("Deleting old subscription ID: %d with URL: %s", sub.ID, sub.CallbackURL)
		if err := s.client.DeleteWebhookSubscription(sub.ID); err != nil {
			log.Printf("Warning: failed to delete subscription %d: %v", sub.ID, err)
		}
	}

	// Create new subscription with current URL
	subscription, err := s.client.CreateWebhookSubscription(callbackURL, verifyToken)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("Subscription already exists, treating as success")
			return nil
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
