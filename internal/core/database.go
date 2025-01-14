package core

import (
	"errors"
	"sync"
)

type Database struct {
	mu    sync.Mutex
	store map[string]string
}

// Create a new database instance
func NewDatabase() *Database {
	return &Database{
		store: make(map[string]string),
	}
}

// Set stores a key-value pair in the database
func (db *Database) Set(key, value string) error {
	// Validate inputs
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if value == "" {
		return errors.New("value cannot be empty")
	}

	// Store the key-value pair
	db.mu.Lock()
	defer db.mu.Unlock()
	db.store[key] = value
	return nil
}

// Get retrieves the value associated with the given key
func (db *Database) Get(key string) (string, error) {
	// Validate input
	if key == "" {
		return "", errors.New("key cannot be empty")
	}

	// Retrieve the value
	db.mu.Lock()
	defer db.mu.Unlock()
	value, exists := db.store[key]
	if !exists {
		return "", errors.New("key not found")
	}
	return value, nil
}
