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

	// Test INCR
	t.Run("INCR command existing integer key", func(t *testing.T) {
		_ = db.Set("counter", "10", 100)

		newValue, err := db.Incr("counter", 5)
		if err != nil || newValue != 15 {
			t.Errorf("expected newValue: 15, got %v (error: %v)", newValue, err)
		}

		value, _ := db.Get("counter")
		if value != "15" {
			t.Errorf("expected value: 15, got %v", value)
		}
	})

	t.Run("INCR command non-existing key", func(t *testing.T) {
		_, err := db.Incr("nonexistent", 3)
		if err == nil || err.Error() != "key not found" {
			t.Errorf("expected error: key not found, got %v", err)
		}
	})

	t.Run("INCR command non-integer value", func(t *testing.T) {
		_ = db.Set("stringValue", "hello", 100)

		_, err := db.Incr("stringValue", 2)
		if err == nil || err.Error() != "value is not an integer" {
			t.Errorf("expected error: value is not an integer, got %v", err)
		}
	})

	// Test Stack and Queue Commands
	t.Run("PUSH command", func(t *testing.T) {
		err := db.Push("list1", "item1")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		err = db.Push("list1", "item2")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("LPOP command", func(t *testing.T) {
		_ = db.Push("list2", "item1")
		_ = db.Push("list2", "item2")
		_ = db.Push("list2", "item3")

		item, err := db.Lpop("list2")
		if err != nil || item != "item1" {
			t.Errorf("expected item1, got %v (error: %v)", item, err)
		}

		item, err = db.Lpop("list2")
		if err != nil || item != "item2" {
			t.Errorf("expected item2, got %v (error: %v)", item, err)
		}
	})

	t.Run("RPOP command", func(t *testing.T) {
		_ = db.Push("list3", "item1")
		_ = db.Push("list3", "item2")
		_ = db.Push("list3", "item3")

		item, err := db.Rpop("list3")
		if err != nil || item != "item3" {
			t.Errorf("expected item3, got %v (error: %v)", item, err)
		}

		item, err = db.Rpop("list3")
		if err != nil || item != "item2" {
			t.Errorf("expected item2, got %v (error: %v)", item, err)
		}
	})

	t.Run("LPOP/RPOP from empty list", func(t *testing.T) {
		_, err := db.Lpop("emptylist")
		if err == nil || err.Error() != "list is empty or does not exist" {
			t.Errorf("expected error: list is empty or does not exist, got %v", err)
		}

		_, err = db.Rpop("emptylist")
		if err == nil || err.Error() != "list is empty or does not exist" {
			t.Errorf("expected error: list is empty or does not exist, got %v", err)
		}
	})
}
