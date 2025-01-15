package core

import (
	"errors"
	"sync"
	"time"
)

type Database struct {
	mu     sync.Mutex
	store  map[string]string
	expiry map[string]int64
}

// Create a new database instance
func NewDatabase() *Database {
	return &Database{
		store:  make(map[string]string),
		expiry: make(map[string]int64),
	}
}

// Set stores a key-value pair in the database
func (db *Database) Set(key, value string, ttlMs int64) error {
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

	if ttlMs > 0 {
		expiryTime := time.Now().UnixMilli() + ttlMs
		db.expiry[key] = expiryTime
	}

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

func (db *Database) StartCleanup(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)

			db.mu.Lock()
			now := time.Now().UnixMilli()
			for key, expiry := range db.expiry {
				if now > expiry {
					delete(db.store, key)
					delete(db.expiry, key)
				}
			}
			db.mu.Unlock()
		}
	}()
}
