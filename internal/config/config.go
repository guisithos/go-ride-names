package config

import "os"

type Config struct {
	StravaAccessToken  string
	StravaRefreshToken string
	StravaClientID     string
	StravaClientSecret string
}

func New() *Config {
	return &Config{
		StravaAccessToken:  os.Getenv("STRAVA_ACCESS_TOKEN"),
		StravaRefreshToken: os.Getenv("STRAVA_REFRESH_TOKEN"),
		StravaClientID:     os.Getenv("STRAVA_CLIENT_ID"),
		StravaClientSecret: os.Getenv("STRAVA_CLIENT_SECRET"),
	}
}
