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
	"Evening Run":               true,
	"Night Run":                 true,
	"Morning Ride":              true,
	"Afternoon Ride":            true,
	"Evening Ride":              true,
	"Night Ride":                true,
	"Morning Walk":              true,
	"Afternoon Walk":            true,
	"Evening Walk":              true,
	"Night Walk":                true,
	"Morning Weight Training":   true,
	"Afternoon Weight Training": true,
	"Lunch Weight Training":     true,
	"Evening Weight Training":   true,
	"Night Weight Training":     true,
	"Morning Swim":              true,
	"Afternoon Swim":            true,
	"Evening Swim":              true,
	"Night Swim":                true,
	"Morning Yoga":              true,
	"Afternoon Yoga":            true,
	"Evening Yoga":              true,
	"Night Yoga":                true,
}

type ActivityService struct {
	client              *strava.Client
	webhookSubscription *strava.WebhookSubscription
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
	// Get the activity details
	activity, err := s.client.GetActivity(activityID)
	if err != nil {
		return err
	}

	// Update the activity name if it's a default name
	return s.UpdateActivityWithFunName(activity)
}

func (s *ActivityService) SubscribeToWebhooks(callbackURL, verifyToken string) error {
	subscription, err := s.client.CreateWebhookSubscription(callbackURL, verifyToken)
	if err != nil {
		return err
	}
	s.webhookSubscription = subscription
	return nil
}
