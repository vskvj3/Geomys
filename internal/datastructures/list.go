package datastructures

import "errors"

type (
	// List represents a doubly linked list.
	List struct {
		head   *node
		tail   *node
		length int
	}

	// Node represents an element in the doubly linked list.
	node struct {
		value interface{}
		prev  *node
		next  *node
	}
)

// New creates a new list.
func NewList() *List {
	return &List{}
}

// LPush adds a value to the left (head) of the list.
func (l *List) LPush(value interface{}) {
	n := &node{value: value}
	if l.length == 0 {
		l.head = n
		l.tail = n
	} else {
		n.next = l.head
		l.head.prev = n
		l.head = n
	}
	l.length++
}

// RPush adds a value to the right (tail) of the list.
func (l *List) RPush(value interface{}) {
	n := &node{value: value}
	if l.length == 0 {
		l.head = n
		l.tail = n
	} else {
		n.prev = l.tail
		l.tail.next = n
		l.tail = n
	}
	l.length++
}

// LPop removes and returns the value from the left (head) of the list.
func (l *List) LPop() (interface{}, error) {
	if l.length == 0 {
		return nil, errors.New("list is empty")
	}
	value := l.head.value
	if l.length == 1 {
		l.head = nil
		l.tail = nil
	} else {
		l.head = l.head.next
		l.head.prev = nil
	}
	l.length--
	return value, nil
}

// RPop removes and returns the value from the right (tail) of the list.
func (l *List) RPop() (interface{}, error) {
	if l.length == 0 {
		return nil, errors.New("list is empty")
	}
	value := l.tail.value
	if l.length == 1 {
		l.head = nil
		l.tail = nil
	} else {
		l.tail = l.tail.prev
		l.tail.next = nil
	}
	l.length--
	return value, nil
}

// Len returns the number of elements in the list.
func (l *List) Len() int {
	return l.length
}

// Size is an alias for Len to maintain consistency with the required operations.
func (l *List) Size() int {
	return l.Len()
}

// Clear removes all elements from the list.
func (l *List) Clear() {
	l.head = nil
	l.tail = nil
	l.length = 0
}
