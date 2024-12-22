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
}

type ActivityService struct {
	client strava.StravaClientInterface
}

func NewActivityService(client strava.StravaClientInterface) *ActivityService {
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

// RenameActivity renames a specific activity with a fun name
func (s *ActivityService) RenameActivity(activityID int64) error {
	// Get activity details
	activity, err := s.client.GetActivity(activityID)
	if err != nil {
		return fmt.Errorf("failed to get activity: %v", err)
	}

	// Only rename if it has a default name
	if !defaultActivityNames[activity.Name] {
		log.Printf("Activity '%s' doesn't have a default name, skipping", activity.Name)
		return nil
	}

	// Use our existing name generation logic
	activityType := getActivityType(activity.Name, activity.SportType)
	newName := getRandomJoke(activityType)

	// Log the name change
	log.Printf("Updating activity name:\n  From: %s\n  Type: %s\n  To:   %s\n",
		activity.Name,
		activityType,
		newName)

	// Update activity name
	if err := s.client.UpdateActivity(activityID, newName); err != nil {
		return fmt.Errorf("failed to update activity: %v", err)
	}

	return nil
}
