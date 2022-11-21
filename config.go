package main

import (
	"encoding/json"
	"os"
)

// Configuration file for the application.
type Config struct{}

// Loads a Config from a given file path.
func ConfigFromJSON(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rsl Config
	json.Unmarshal(raw, &rsl)
	if err != nil {
		return nil, err
	}
	return &rsl, nil
}

// Saves a Config instance in JSON file.
func (c Config) ToJSON(path string) error {
	dt, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, dt, 0644)
}
