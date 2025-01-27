/*
This file only exists as a reference
- Currently using list(Doubly Linked list) can be found as /core/list.go
*/
package datastructures

import (
	"errors"
)

// Deque represents a double-ended queue.
type Deque[T any] struct {
	data     []T
	size     int
	head     int
	tail     int
	capacity int
}

// NewDeque creates a new Deque with the specified capacity.
func NewDeque[T any](capacity int) *Deque[T] {
	if capacity <= 0 {
		panic("capacity must be greater than 0")
	}
	return &Deque[T]{
		data:     make([]T, capacity),
		capacity: capacity,
		head:     0,
		tail:     capacity - 1,
	}
}

// PushFront adds an element to the front of the deque.
func (d *Deque[T]) PushFront(value T) error {
	if d.size == d.capacity {
		return errors.New("deque is full")
	}
	d.head = (d.head - 1 + d.capacity) % d.capacity
	d.data[d.head] = value
	d.size++
	return nil
}

// PushBack adds an element to the back of the deque.
func (d *Deque[T]) PushBack(value T) error {
	if d.size == d.capacity {
		return errors.New("deque is full")
	}
	d.tail = (d.tail + 1) % d.capacity
	d.data[d.tail] = value
	d.size++
	return nil
}

// PopFront removes an element from the front of the deque.
func (d *Deque[T]) PopFront() (T, error) {
	if d.size == 0 {
		var zeroValue T
		return zeroValue, errors.New("deque is empty")
	}
	value := d.data[d.head]
	d.head = (d.head + 1) % d.capacity
	d.size--
	return value, nil
}

// PopBack removes an element from the back of the deque.
func (d *Deque[T]) PopBack() (T, error) {
	if d.size == 0 {
		var zeroValue T
		return zeroValue, errors.New("deque is empty")
	}
	value := d.data[d.tail]
	d.tail = (d.tail - 1 + d.capacity) % d.capacity
	d.size--
	return value, nil
}

// Front returns the element at the front of the deque.
func (d *Deque[T]) Front() (T, error) {
	if d.size == 0 {
		var zeroValue T
		return zeroValue, errors.New("deque is empty")
	}
	return d.data[d.head], nil
}

// Back returns the element at the back of the deque.
func (d *Deque[T]) Back() (T, error) {
	if d.size == 0 {
		var zeroValue T
		return zeroValue, errors.New("deque is empty")
	}
	return d.data[d.tail], nil
}

// Size returns the number of elements in the deque.
func (d *Deque[T]) Size() int {
	return d.size
}

// Empty checks if the deque is empty.
func (d *Deque[T]) Empty() bool {
	return d.size == 0
}
