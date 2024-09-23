package config

import (
	"encoding/json"
	"os"
)

// ServerConfig represents the server-related configurations
type ServerConfig struct {
	Port       int    `json:"port"`
	TLSEnabled bool   `json:"tls_enabled"`
	CertFile   string `json:"tls_cert_file"`
	KeyFile    string `json:"tls_key_file"`
}

// CommandConfig represents the configuration for allowed commands
type CommandConfig struct {
	Allowed []string `json:"allowed"`
}

type Route struct {
	Path    string `json:"path"`
	Method  string `json:"method"`
	Handler string `json:"handler"`
	Action  string `json:"action"`
}

type UI struct {
	Enabled bool `json:"enabled"`
}

// AppConfig holds the full application configuration
type AppConfig struct {
	Server           ServerConfig  `json:"server"`
	Commands         CommandConfig `json:"commands"`
	Routes           []Route       `json:"routes"`
	UI               UI            `json:"UI"`
	EncryptionConfig struct {
		Enabled bool `json:"enabled"` // Add this field to toggle encryption
	} `json:"encryption"`
}

var GlobalConfig *AppConfig

// LoadAppConfig loads the JSON configuration from a file
func LoadAppConfig(path string) (*AppConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &AppConfig{}
	GlobalConfig = &AppConfig{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
