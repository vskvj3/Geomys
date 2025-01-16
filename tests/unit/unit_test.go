package unit

import (
	"testing"
	"time"

	"github.com/vskvj3/geomys/internal/core"
)

func TestCoreCommands(t *testing.T) {
	db := core.NewDatabase()
	db.StartCleanup(100 * time.Millisecond)

	// Test SET
	t.Run("SET command", func(t *testing.T) {
		err := db.Set("key1", "value1", 100)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	// Test GET
	t.Run("GET command existing key", func(t *testing.T) {
		_ = db.Set("key1", "value1", 100)
		value, err := db.Get("key1")
		if err != nil || value != "value1" {
			t.Errorf("expected value1, got %v (exists: %v)", value, err)
		}
	})

	t.Run("GET command non-existing key", func(t *testing.T) {
		_, err := db.Get("nonexistent")
		if err == nil {
			t.Errorf("expected key to not exist")
		}
	})

	// Test Expiry
	t.Run("SET with expiry", func(t *testing.T) {
		db.Set("tempkey", "tempvalue", 500) // 500ms expiry
		time.Sleep(300 * time.Millisecond)
		value, err := db.Get("tempkey")
		if err != nil || value != "tempvalue" {
			t.Errorf("expected tempvalue, got %v (err: %v)", value, err)
		}

		time.Sleep(300 * time.Millisecond)
		_, err = db.Get("tempkey")
		if err == nil {
			t.Errorf("expected key to expire")
		}
	})

	t.Run("INCR command existing integer key", func(t *testing.T) {
		// Set an initial integer value for the key
		_ = db.Set("counter", "10", 100)

		// Increment the value by 5
		newValue, err := db.Incr("counter", 5)
		if err != nil || newValue != 15 {
			t.Errorf("expected newValue: 15, got %v (error: %v)", newValue, err)
		}

		// Verify the value in the database
		value, _ := db.Get("counter")
		if value != "15" {
			t.Errorf("expected value: 15, got %v", value)
		}
	})

	t.Run("INCR command non-existing key", func(t *testing.T) {
		// Attempt to increment a non-existing key
		_, err := db.Incr("nonexistent", 3)
		if err == nil || err.Error() != "key not found" {
			t.Errorf("expected error: key not found, got %v", err)
		}
	})

	t.Run("INCR command non-integer value", func(t *testing.T) {
		// Set a non-integer value for the key
		_ = db.Set("stringValue", "hello", 100)

		// Attempt to increment the non-integer value
		_, err := db.Incr("stringValue", 2)
		if err == nil || err.Error() != "value is not an integer" {
			t.Errorf("expected error: value is not an integer, got %v", err)
		}
	})

}
