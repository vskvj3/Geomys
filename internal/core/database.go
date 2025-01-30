package core

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/vskvj3/geomys/internal/persistence"
)

type Database struct {
	mu     sync.Mutex
	store  map[string]string
	expiry map[string]int64
	lists  map[string]*List
}

// Create a new database instance
func NewDatabase() *Database {
	return &Database{
		store:  make(map[string]string),
		expiry: make(map[string]int64),
		lists:  make(map[string]*List),
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

	leftValue, err := list.LPop()
	if err != nil {
		return 0, err
	}

	// return the leftmost value
	return leftValue, nil
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

	rightValue, err := list.RPop()
	if err != nil {
		return 0, err
	}

	// return the rightmost value
	return rightValue, nil
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

// rebuild database at the run time
func (db *Database) RebuildFromPersistence() error {
	// Load stored requests
	disk, err := persistence.CreateOrReplacePersistence()
	if err != nil {
		return err
	}
	requests, err := disk.LoadRequests()
	if err != nil {
		return err
	}
	fmt.Println("requests", requests)

	// Replay each request using correct database function
	for _, req := range requests {
		switch req["command"] {
		case "SET":
			db.Set(req["key"].(string), req["value"].(string), 0) // Assuming no TTL
		case "INCR":
			key := req["key"].(string)
			offset, err := strconv.Atoi(req["offset"].(string))
			if err != nil {
				fmt.Println("Offset conversion failed in databse rebuild")
				break
			}
			db.Incr(key, offset)
		case "PUSH":
			db.Push(req["key"].(string), req["value"])
		case "RPOP":
			db.Rpop(req["key"].(string))
		case "LPOP":
			db.Lpop(req["key"].(string))
		default:
			continue
		}
	}

	return nil
}
