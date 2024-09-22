package config

import (
	"encoding/json"
	"os"
)

// Route represents a single API route
type Route struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Action string `json:"action"`
}

// Config represents the structure of the JSON config
type Config struct {
	Routes []Route `json:"routes"`
}

// LoadConfig loads the JSON configuration from the file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into the Config struct
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
