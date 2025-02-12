package utils

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

// Config struct holds application configuration
type Config struct {
	InternalPort  int    `json:"internal_port"`
	ExternalPort  int    `json:"external_port"`
	DefaultExpiry int    `json:"default_expiry"`
	Persistence   string `json:"persistence"`
	Replication   bool   `json:"replication_enabled"`
	NodeID        int    `json:"node_id"`
	IsLeader      bool   `json:"leader_id"`
	Sharding      bool   `json:"sharding_enabled"`
	ClusterMode   bool   `json:"cluster_mode"`
}

var (
	configInstance *Config   // Singleton configInstance
	configOnce     sync.Once // Ensures thread-safe initialization
)

// LoadConfig initializes the singleton configInstance
func LoadConfig(filename string) error {
	var err error
	configOnce.Do(func() {
		configInstance, err = loadConfigFromFile(filename)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	})
	return err
}

// loadConfigFromFile reads and parses the config file
func loadConfigFromFile(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
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

	applyDefaults(config)
	return config, nil
}

// GetConfig returns the singleton config configInstance
func GetConfig() (*Config, error) {
	if configInstance == nil {
		return nil, errors.New("Config not initialized. Call LoadConfig() first")
	}
	return configInstance, nil
}

// getDefaultConfig returns default config values
func getDefaultConfig() *Config {
	return &Config{
		InternalPort:  6379,
		DefaultExpiry: 60000,
		Persistence:   "bufferedwrite",
		Replication:   false,
		Sharding:      false,
		IsLeader:      false,
	}
}

// applyDefaults ensures missing values get defaults
func applyDefaults(config *Config) {
	if config.InternalPort == 0 {
		config.InternalPort = 6379
	}
	if config.DefaultExpiry == 0 {
		config.DefaultExpiry = 60000
	}
	if config.Persistence != "writethroughdisk" && config.Persistence != "bufferedwrite" {
		config.Persistence = "writethroughdisk"
	}
}
