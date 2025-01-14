package core

import "sync"

type Database struct {
	mu    sync.Mutex
	store map[string]string
}

func NewDatabase() *Database {
	return &Database{
		store: make(map[string]string),
	}
}

func (db *Database) Set(key, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.store[key] = value
}

func (db *Database) Get(key string) (string, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	value, exists := db.store[key]
	return value, exists
}
