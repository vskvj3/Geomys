package core

import (
	"errors"
	"strconv"
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

// Incr increments the integer value of a key by a given offset
func (db *Database) Incr(key string, offset int) (int, error) {
	// Validate input
	if key == "" {
		return 0, errors.New("key cannot be empty")
	}

	// Lock the database to ensure thread safety
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check if the key exists
	value, exists := db.store[key]
	if !exists {
		return 0, errors.New("key not found")
	}

	// Parse the current value as an integer
	currentValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("value is not an integer")
	}

	// Increment the value by the given offset
	newValue := currentValue + offset
	db.store[key] = strconv.Itoa(newValue)

	return newValue, nil
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
