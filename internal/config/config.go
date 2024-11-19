package config

import "os"

type Config struct {
	StravaAccessToken string
}

func New() *Config {
	return &Config{
		StravaAccessToken: os.Getenv("STRAVA_ACCESS_TOKEN"),
	}
}
