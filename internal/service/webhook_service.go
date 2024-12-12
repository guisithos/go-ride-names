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

	for _, sub := range subscriptions {
		if sub.CallbackURL == callbackURL {
			log.Printf("Found existing subscription with ID: %d", sub.ID)
			s.webhookSubscription = &sub
			return nil
		}
	}

	subscription, err := s.client.CreateWebhookSubscription(callbackURL, verifyToken)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("Subscription already exists, treating as success")
			return nil
		}
		return fmt.Errorf("failed to create subscription: %v", err)
	}

	log.Printf("Successfully created new webhook subscription: ID=%d", subscription.ID)
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
