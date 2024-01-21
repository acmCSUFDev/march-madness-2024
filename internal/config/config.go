package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type HackathonConfig struct {
	StartTime time.Time `json:"start_time"`
	Duration  Duration  `json:"duration"`
}

func (c HackathonConfig) EndTime() time.Time {
	return c.StartTime.Add(c.Duration.Duration())
}

func (c HackathonConfig) IsOpen(now time.Time) bool {
	return c.StartTime.Before(now) && c.EndTime().After(now)
}

type ProblemsConfig struct {
	PWD      string   `json:"pwd"`
	Paths    []string `json:"paths"`
	Schedule struct {
		Start time.Time `json:"start"`
		Every Duration  `json:"every"`
	} `json:"schedule"`
	Cooldown Duration `json:"cooldown"`
}

type Config struct {
	HTTPAddress string `json:"http_address"`
	Paths       struct {
		Frontend      string `json:"frontend"`
		Database      string `json:"database"`
		ProblemsCache string `json:"problems_cache"`
		SecretKey     string `json:"secret_key"`
	} `json:"paths"`
	Problems             ProblemsConfig  `json:"problems"`
	Hackathon            HackathonConfig `json:"hackathon"`
	OpenRegistrationTime time.Time       `json:"open_registration_time"`
}

// ParseFile parses the config file at the given path.
func ParseFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &config, nil
}
