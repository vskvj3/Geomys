package core

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vskvj3/geomys/internal/persistence"
)

type Database struct {
	mu          sync.Mutex
	store       map[string]string
	expiry      map[string]int64
	lists       map[string]*List
	persistence *persistence.Persistence
}

// Create a new database instance
func NewDatabase(persistence *persistence.Persistence) *Database {
	return &Database{
		store:       make(map[string]string),
		expiry:      make(map[string]int64),
		lists:       make(map[string]*List),
		persistence: persistence,
	}
}

// Set stores a key-value pair in the database
func (db *Database) Set(key string, value string, ttlMs int64) error {
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

	// Log the operation
	operation := "SET " + key + " " + value + " " + strconv.FormatInt(ttlMs, 10)
	if err := db.persistence.LogOperation(operation); err != nil {
		return err
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

// LPush adds an item to the left of the list
func (db *Database) LPush(key string, value interface{}) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// Initialize the list if it doesn't exist
	if _, exists := db.lists[key]; !exists {
		db.lists[key] = NewList()
	}

	// Add the value to the left of the list
	db.lists[key].LPush(value)
	return nil
}

// RPush adds an item to the right of the list
func (db *Database) Push(key string, value interface{}) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	// Initialize the list if it doesn't exist
	if _, exists := db.lists[key]; !exists {
		db.lists[key] = NewList()
	}

	// Add the value to the right of the list
	db.lists[key].RPush(value)
	return nil
}

// LPop removes and returns the item from the left of the list
func (db *Database) Lpop(key string) (interface{}, error) {
	if key == "" {
		return nil, errors.New("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	list, exists := db.lists[key]
	if !exists {
		return nil, errors.New("list does not exist")
	}

	// Remove and return the leftmost value
	return list.LPop()
}

// RPop removes and returns the item from the right of the list
func (db *Database) Rpop(key string) (interface{}, error) {
	if key == "" {
		return nil, errors.New("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	list, exists := db.lists[key]
	if !exists {
		return nil, errors.New("list does not exist")
	}

	// Remove and return the rightmost value
	return list.RPop()
}

// Len returns the length of a list
func (db *Database) Len(key string) (int, error) {
	if key == "" {
		return 0, errors.New("key cannot be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	list, exists := db.lists[key]
	if !exists {
		return 0, errors.New("list does not exist")
	}

	return list.Len(), nil
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

func (db *Database) RebuildFromPersistence() error {
	operations, err := db.persistence.LoadOperations()
	if err != nil {
		return err
	}

	for _, op := range operations {
		parts := strings.Split(op, " ")
		if len(parts) < 2 {
			continue
		}

		command := parts[0]
		switch command {
		case "SET":
			if len(parts) >= 4 {
				key := parts[1]
				value := parts[2]
				ttlMs, _ := strconv.ParseInt(parts[3], 10, 64)
				db.Set(key, value, ttlMs)
			}
			// Add cases for other commands (INCR, LPush, etc.)
		}
	}
	return nil
}
