package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddress string `json:"http_address"`
	Paths       struct {
		Frontend  string `json:"frontend"`
		Database  string `json:"database"`
		SecretKey string `json:"secret_key"`
	} `json:"paths"`
	Problems             ProblemsConfig  `json:"problems"`
	Hackathon            HackathonConfig `json:"hackathon"`
	OpenRegistrationTime time.Time       `json:"open_registration_time"`
}

type ProblemsConfig struct {
	Modules  []ProblemModule `json:"modules"`
	Schedule struct {
		Start time.Time `json:"start"`
		Every Duration  `json:"every"`
	} `json:"schedule"`
	Cooldown Duration `json:"cooldown"`
}

type ProblemModule struct {
	Command string `json:"cmd"`
	README  string `json:"readme"`
}

type HackathonConfig struct {
	StartTime time.Time `json:"start_time"`
	Duration  Duration  `json:"duration"`
	Location  string    `json:"location"`
}

func (c HackathonConfig) EndTime() time.Time {
	return c.StartTime.Add(c.Duration.Duration())
}

func (c HackathonConfig) IsOpen(now time.Time) bool {
	return c.StartTime.Before(now) && c.EndTime().After(now)
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
