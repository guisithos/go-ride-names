package service

import "github.com/guisithos/go-ride-names/strava"

type ActivityService struct {
	client *strava.Client
}

func NewActivityService(client *strava.Client) *ActivityService {
	return &ActivityService{
		client: client,
	}
}

func (s *ActivityService) GetAuthenticatedAthlete() (*strava.Athlete, error) {
	return s.client.GetAuthenticatedAthlete()
}
