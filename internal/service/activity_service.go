package service

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/guisithos/go-ride-names/internal/strava"
)

var defaultActivityNames = map[string]bool{
	"Morning Run":               true,
	"Afternoon Run":             true,
	"Lunch Run":                 true,
	"Evening Run":               true,
	"Night Run":                 true,
	"Morning Ride":              true,
	"Afternoon Ride":            true,
	"Lunch Ride":                true,
	"Evening Ride":              true,
	"Night Ride":                true,
	"Morning Walk":              true,
	"Afternoon Walk":            true,
	"Lunch Walk":                true,
	"Evening Walk":              true,
	"Night Walk":                true,
	"Morning Weight Training":   true,
	"Afternoon Weight Training": true,
	"Lunch Weight Training":     true,
	"Evening Weight Training":   true,
	"Night Weight Training":     true,
	"Morning Swim":              true,
	"Afternoon Swim":            true,
	"Lunch Swim":                true,
	"Evening Swim":              true,
	"Night Swim":                true,
	"Morning Yoga":              true,
	"Afternoon Yoga":            true,
	"Lunch Yoga":                true,
	"Evening Yoga":              true,
	"Night Yoga":                true,
	"Morning Workout":           true,
	"Afternoon Workout":         true,
	"Lunch Workout":             true,
	"Evening Workout":           true,
	"Night Workout":             true,
}

type ActivityService struct {
	client         *strava.Client
	webhookManager *WebhookManager
}

func NewActivityService(client *strava.Client) *ActivityService {
	return &ActivityService{
		client: client,
	}
}

func (s *ActivityService) GetAuthenticatedAthlete() (*strava.Athlete, error) {
	return s.client.GetAuthenticatedAthlete()
}

func (s *ActivityService) ListActivities(page, perPage int, before, after int64, updateNames bool) ([]strava.Activity, error) {
	activities, err := s.client.GetAthleteActivities(page, perPage, before, after)
	if err != nil {
		return nil, fmt.Errorf("error getting activities: %v", err)
	}

	if updateNames {
		for i := range activities {
			if err := s.UpdateActivityWithFunName(&activities[i]); err != nil {
				log.Printf("Warning: failed to update activity %d: %v", activities[i].ID, err)
			}
		}
	}

	return activities, nil
}

func (s *ActivityService) UpdateActivityWithFunName(activity *strava.Activity) error {
	// Check if the activity has a default name
	if !defaultActivityNames[activity.Name] {
		return nil // Not a default name, no need to update
	}

	// Get activity type using both name and sport_type
	activityType := getActivityType(activity.Name, activity.SportType)
	joke := getRandomJoke(activityType)

	// Log the name change
	fmt.Printf("Updating activity name:\n  From: %s\n  Type: %s\n  To:   %s\n\n",
		activity.Name,
		activityType,
		joke)

	// Update the activity name
	if err := s.client.UpdateActivity(activity.ID, joke); err != nil {
		return fmt.Errorf("error updating activity: %v", err)
	}

	// Update the local activity name
	activity.Name = joke
	return nil
}

// Create a package-level random number generator
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func getRandomJoke(activityType string) string {
	jokes, exists := activityJokes[activityType]
	if !exists || len(jokes) == 0 {
		jokes = activityJokes[Default]
	}
	return jokes[rng.Intn(len(jokes))]
}

func (s *ActivityService) ProcessNewActivity(activityID int64) error {
	activity, err := s.client.GetActivity(activityID)
	if err != nil {
		return fmt.Errorf("error getting activity: %v", err)
	}

	// Only process if it has a default name
	if defaultActivityNames[activity.Name] {
		return s.UpdateActivityWithFunName(activity)
	}

	return nil
}

// InitializeWebhooks sets up the webhook manager with the provided callback URL and verify token
func (s *ActivityService) InitializeWebhooks(callbackURL, verifyToken string) error {
	s.webhookManager = NewWebhookManager(s.client, callbackURL, verifyToken)
	return s.webhookManager.Start()
}

func (s *ActivityService) GetWebhookStatus() (bool, time.Time) {
	if s.webhookManager == nil {
		return false, time.Time{}
	}
	return s.webhookManager.GetStatus()
}

func (s *ActivityService) SubscribeToWebhooks(callbackURL, verifyToken string) error {
	if s.webhookManager == nil {
		if err := s.InitializeWebhooks(callbackURL, verifyToken); err != nil {
			return fmt.Errorf("failed to initialize webhook manager: %v", err)
		}
		return nil
	}
	return s.webhookManager.ForceCheck()
}

func (s *ActivityService) UnsubscribeFromWebhooks() error {
	if s.webhookManager == nil || s.webhookManager.subscription == nil {
		return nil
	}

	if err := s.client.DeleteWebhookSubscription(s.webhookManager.subscription.ID); err != nil {
		return fmt.Errorf("error deleting webhook subscription: %v", err)
	}

	s.webhookManager = nil
	return nil
}
