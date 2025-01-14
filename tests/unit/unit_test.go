package unit

import (
	"testing"

	"github.com/vskvj3/geomys/internal/core"
)

func TestCoreCommands(t *testing.T) {
	db := core.NewDatabase()

	// Test SET
	t.Run("SET command", func(t *testing.T) {
		err := db.Set("key1", "value1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	// Test GET
	t.Run("GET command existing key", func(t *testing.T) {
		_ = db.Set("key1", "value1")
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
}
