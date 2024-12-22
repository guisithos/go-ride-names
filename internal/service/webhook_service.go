package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/guisithos/go-ride-names/internal/strava"
)

type WebhookService struct {
	client *strava.Client
}

func NewWebhookService(client *strava.Client) *WebhookService {
	return &WebhookService{
		client: client,
	}
}

func (s *WebhookService) SubscribeToWebhooks(callbackURL, verifyToken string) error {
	log.Printf("Starting webhook subscription process for URL: %s", callbackURL)

	if err := s.cleanupExistingSubscriptions(); err != nil {
		return fmt.Errorf("cleanup failed: %v", err)
	}

	subscription, err := s.client.CreateWebhookSubscription(callbackURL, verifyToken)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("Subscription already exists")
			return nil
		}
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	log.Printf("Successfully created webhook subscription: ID=%d", subscription.ID)
	return nil
}

func (s *WebhookService) cleanupExistingSubscriptions() error {
	subscriptions, err := s.client.ListWebhookSubscriptions()
	if err != nil {
		return fmt.Errorf("error listing subscriptions: %v", err)
	}

	for _, sub := range subscriptions {
		if err := s.client.DeleteWebhookSubscription(sub.ID); err != nil {
			log.Printf("Warning: failed to delete subscription %d: %v", sub.ID, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
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
