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
	expiry map[string]int64    // when the key expires?
	lists  map[string][]string // Key-list mappings (used for both stacks and queues)
}

// Create a new database instance
func NewDatabase() *Database {
	return &Database{
		store:  make(map[string]string),
		expiry: make(map[string]int64),
		lists:  make(map[string][]string),
	}
}

// Set stores a key-value pair in the database
func (db *Database) Set(key, value string, ttlMs int64) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if value == "" {
		return errors.New("value cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	db.store[key] = value
	if ttlMs > 0 {
		db.expiry[key] = time.Now().UnixMilli() + ttlMs
	}
	return nil
}

// Get retrieves the value associated with the given key
func (db *Database) Get(key string) (string, error) {
	if key == "" {
		return "", errors.New("key cannot be empty")
	}

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
	if key == "" {
		return 0, errors.New("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	value, exists := db.store[key]
	if !exists {
		return 0, errors.New("key not found")
	}

	currentValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("value is not an integer")
	}

	newValue := currentValue + offset
	db.store[key] = strconv.Itoa(newValue)

	return newValue, nil
}

// Push adds an item to the stack/queue
func (db *Database) Push(key, value string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// Initialize the list if it doesn't exist
	if _, exists := db.lists[key]; !exists {
		db.lists[key] = []string{}
	}

	// Append the value to the list
	db.lists[key] = append(db.lists[key], value)
	return nil
}

// Lpop removes and returns the item from the left of the list
func (db *Database) Lpop(key string) (string, error) {
	if key == "" {
		return "", errors.New("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	list, exists := db.lists[key]
	if !exists || len(list) == 0 {
		return "", errors.New("list is empty or does not exist")
	}

	// Retrieve and remove the first element
	value := list[0]
	db.lists[key] = list[1:]

	return value, nil
}

// Rpop removes and returns the item from the right of the list
func (db *Database) Rpop(key string) (string, error) {
	if key == "" {
		return "", errors.New("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	list, exists := db.lists[key]
	if !exists || len(list) == 0 {
		return "", errors.New("list is empty or does not exist")
	}

	// Retrieve and remove the last element
	value := list[len(list)-1]
	db.lists[key] = list[:len(list)-1]

	return value, nil
}

// StartCleanup starts a background goroutine to clean up expired keys
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
