package strava

import "time"

type Athlete struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	City      string `json:"city"`
	Country   string `json:"country"`
}

type Activity struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Distance       float64   `json:"distance"`
	MovingTime     int       `json:"moving_time"`
	Type           string    `json:"type"`
	SportType      string    `json:"sport_type"`
	StartDate      time.Time `json:"start_date"`
	StartDateLocal time.Time `json:"start_date_local"`
}
