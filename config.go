package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddress string `json:"http_address"`
	Paths       struct {
		Frontend      string `json:"frontend"`
		Database      string `json:"database"`
		ProblemsCache string `json:"problems_cache"`
		SecretKey     string `json:"secret_key"`
	} `json:"paths"`
	Problems struct {
		PWD      string   `json:"pwd"`
		Paths    []string `json:"paths"`
		Schedule struct {
			Start time.Time      `json:"start"`
			Every configDuration `json:"every"`
		} `json:"schedule"`
	} `json:"problems"`
}

func parseConfig(path string) (*Config, error) {
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

type configDuration time.Duration

func (d *configDuration) UnmarshalText(b []byte) error {
	dur, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = configDuration(dur)
	return nil
}

func (d configDuration) Duration() time.Duration {
	return time.Duration(d)
}
