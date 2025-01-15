package utils

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	Port          int  `json:"port"`                // Server port
	DefaultExpiry int  `json:"default_expiry"`      // Default expiration time in milliseconds
	Replication   bool `json:"replication_enabled"` // Enable/disable replication
	Sharding      bool `json:"sharding_enabled"`    // Enable/disable sharding
}

// LoadConfig reads the configuration from the given file
func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		// If the file doesn't exist, return defaults
		if errors.Is(err, os.ErrNotExist) {
			return getDefaultConfig(), nil
		}
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	// Validate and apply defaults if necessary
	applyDefaults(config)
	return config, nil
}

// getDefaultConfig returns a Config struct with default values
func getDefaultConfig() *Config {
	return &Config{
		Port:          6379,
		DefaultExpiry: 60000, // 60 seconds
		Replication:   false,
		Sharding:      false,
	}
}

// applyDefaults ensures that missing or zero-value fields get default values
func applyDefaults(config *Config) {
	if config.Port == 0 {
		config.Port = 6379
	}
	if config.DefaultExpiry == 0 {
		config.DefaultExpiry = 60000 // 60 seconds
	}
}
