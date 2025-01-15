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

}
