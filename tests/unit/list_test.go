package unit

import (
	"testing"

	"github.com/vskvj3/geomys/internal/datastructures"
)

func TestLPush(t *testing.T) {
	l := datastructures.NewList()
	l.LPush("value1")
	if l.Len() != 1 {
		t.Errorf("expected length 1, got %d", l.Len())
	}
}

func TestRPush(t *testing.T) {
	l := datastructures.NewList()
	l.RPush("value1")
	if l.Len() != 1 {
		t.Errorf("expected length 1, got %d", l.Len())
	}
}

func TestLPop(t *testing.T) {
	l := datastructures.NewList()
	l.LPush("value1")
	l.LPush("value2")
	val, err := l.LPop()
	if err != nil {
		t.Errorf("failed %v", err.Error())
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}
	if l.Len() != 1 {
		t.Errorf("expected length 1, got %d", l.Len())
	}
}

func TestRPop(t *testing.T) {
	l := datastructures.NewList()
	l.RPush("value1")
	l.RPush("value2")
	val, err := l.RPop()
	if err != nil {
		t.Errorf("failed %v", err.Error())
	}
	if val != "value2" {
		t.Errorf("expected value2, got %v", val)
	}
	if l.Len() != 1 {
		t.Errorf("expected length 1, got %d", l.Len())
	}
}
