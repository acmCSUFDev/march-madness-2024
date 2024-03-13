package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"dev.acmcsuf.com/march-madness-2024/internal/config"
	"dev.acmcsuf.com/march-madness-2024/server"
)

type Config struct {
	HTTPAddress string `json:"http_address"`
	Paths       struct {
		Frontend      string `json:"frontend"`
		Database      string `json:"database"`
		ProblemsCache string `json:"problems_cache"`
		SecretKey     string `json:"secret_key"`
	} `json:"paths"`
	Problems             ProblemsConfig         `json:"problems"`
	Hackathon            server.HackathonConfig `json:"hackathon"`
	OpenRegistrationTime time.Time              `json:"open_registration_time"`
}

type ProblemsConfig struct {
	Modules  []ProblemModule `json:"modules"`
	Schedule struct {
		Start time.Time       `json:"start"`
		Every config.Duration `json:"every"`
	} `json:"schedule"`
	Cooldown config.Duration `json:"cooldown"`
}

type ProblemModule struct {
	Command string `json:"cmd"`
	README  string `json:"readme"`
}

// ParseConfigFile parses the config file at the given path.
func ParseConfigFile(path string) (*Config, error) {
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
