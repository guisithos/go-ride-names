package service

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/guisithos/go-ride-names/internal/strava"
)

type WebhookManager struct {
	client        *strava.Client
	callbackURL   string
	verifyToken   string
	subscription  *strava.WebhookSubscription
	healthStatus  bool
	lastCheck     time.Time
	checkInterval time.Duration
	mu            sync.RWMutex
}

func NewWebhookManager(client *strava.Client, callbackURL, verifyToken string) *WebhookManager {
	return &WebhookManager{
		client:        client,
		callbackURL:   callbackURL,
		verifyToken:   verifyToken,
		checkInterval: 15 * time.Minute, // Default check interval
		healthStatus:  false,
	}
}

func (wm *WebhookManager) Start() error {
	// Initial subscription check
	if err := wm.ensureSubscription(); err != nil {
		return fmt.Errorf("initial subscription failed: %v", err)
	}

	// Start periodic health checks
	go wm.startPeriodicHealthCheck()
	return nil
}

func (wm *WebhookManager) ensureSubscription() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	log.Printf("Checking webhook subscription status...")

	// List existing subscriptions
	subs, err := wm.client.ListWebhookSubscriptions()
	if err != nil {
		log.Printf("Error listing subscriptions: %v", err)
		wm.healthStatus = false
		return err
	}

	// Check for existing valid subscription
	for _, sub := range subs {
		if sub.CallbackURL == wm.callbackURL {
			log.Printf("Found existing subscription (ID: %d)", sub.ID)
			wm.subscription = &sub
			wm.healthStatus = true
			wm.lastCheck = time.Now()
			return nil
		}
	}

	// Create new subscription if none exists
	log.Printf("No existing subscription found, creating new one...")
	sub, err := wm.client.CreateWebhookSubscription(wm.callbackURL, wm.verifyToken)
	if err != nil {
		log.Printf("Failed to create subscription: %v", err)
		wm.healthStatus = false
		return err
	}

	log.Printf("Successfully created new subscription (ID: %d)", sub.ID)
	wm.subscription = sub
	wm.healthStatus = true
	wm.lastCheck = time.Now()
	return nil
}

func (wm *WebhookManager) startPeriodicHealthCheck() {
	ticker := time.NewTicker(wm.checkInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := wm.checkHealth(); err != nil {
			log.Printf("Webhook health check failed: %v", err)
			// Try to fix the subscription
			if err := wm.ensureSubscription(); err != nil {
				log.Printf("Failed to fix subscription: %v", err)
			}
		}
	}
}

func (wm *WebhookManager) checkHealth() error {
	wm.mu.RLock()
	sub := wm.subscription
	wm.mu.RUnlock()

	if sub == nil {
		return fmt.Errorf("no active subscription")
	}

	// Verify subscription is still valid
	subs, err := wm.client.ListWebhookSubscriptions()
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %v", err)
	}

	for _, s := range subs {
		if s.ID == sub.ID {
			wm.mu.Lock()
			wm.healthStatus = true
			wm.lastCheck = time.Now()
			wm.mu.Unlock()
			return nil
		}
	}

	return fmt.Errorf("subscription not found")
}

func (wm *WebhookManager) GetStatus() (bool, time.Time) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.healthStatus, wm.lastCheck
}

func (wm *WebhookManager) ForceCheck() error {
	return wm.ensureSubscription()
}
