package ringbuf

import (
	"github.com/bdreece/gollections/errors"
	"github.com/bdreece/gollections/iterator"
)

// RingBuf is the ring buffer data structure
type RingBuf[T any] struct {
	data     []T
	capacity int
	length   int
	head     int
	tail     int
}

// New constructs a new RingBuf with capacity
func New[T any](capacity int) *RingBuf[T] {
	return &RingBuf[T]{
		data:     make([]T, capacity),
		capacity: capacity,
		length:   0,
		head:     0,
		tail:     0,
	}
}

// Dequeue reads an item from the RingBuf,
// advancing the head pointer. Returns
// nil, errors.Empty if ring buffer is
// empty.
func (b *RingBuf[T]) Dequeue() (*T, error) {
	if b.length <= 0 {
		return nil, errors.Empty{}
	}

	val := new(T)
	*val = b.data[b.head]

	b.head = (b.head + 1) % b.capacity
	b.length -= 1
	return val, nil
}

// Peek reads an item from the RingBuf
// without advancing the head pointer,
// allowing the value to be read again.
// Returns nil, errors.Empty if ring
// buffer is empty.
func (b RingBuf[T]) Peek() (*T, error) {
	if b.length <= 0 {
		return nil, errors.Empty{}
	}
	val := new(T)
	*val = b.data[b.head]
	return val, nil
}

// Collect writes a variable number of items
// into the RingBuf. This method implements
// part of the Iterator interface.
func (b *RingBuf[T]) Collect(values ...T) {
	for _, value := range values {
		b.Enqueue(value)
	}
}

// IntoIterator returns an iterator over the items
// in the RingBuf. This method implements part
// of the Iterable interface.
func (b *RingBuf[T]) IntoIterator() iterator.Iterator[T] {
	return &Iterator[T]{b}
}

// Enqueue writes an item into the RingBuf,
// advancing the tail pointer.
func (b *RingBuf[T]) Enqueue(item T) {
	b.data[b.tail] = item
	b.tail = (b.tail + 1) % b.capacity
	if b.length < b.capacity {
		b.length += 1
	}
}

// Clear reconstructs the RingBuf in place,
// effectively zeroing all the items.
func (b *RingBuf[T]) Clear() {
	b.data = make([]T, b.capacity)
}

// Iterator provides an iterator over the
// items in a RingBuf
type Iterator[T any] struct {
	*RingBuf[T]
}

// Next returns the next item in the RingBuf.
// Returns nil, errors.Empty after the last
// item has been read.
func (iter *Iterator[T]) Next() (*T, error) {
	return iter.RingBuf.Dequeue()
}
