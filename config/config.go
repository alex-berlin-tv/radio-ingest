package config

import (
	"encoding/json"
	"os"
)

// Configuration file for the application.
type Config struct {
	DomainId      string `json:"domain_id"`
	ApiSecret     string `json:"api_secret"`
	SessionId     string `json:"session_id"`
	Port          int    `json:"port"`
	StackfieldURL string `json:"stackfield_url"`
}

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

// Returns a Config instance with default values.
func ConfigFromDefaults() Config {
	return Config{}
}

// Saves a Config instance in JSON file.
func (c Config) ToJSON(path string) error {
	dt, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, dt, 0644)
}
