package ringbuf

import "errors"

var ErrEmpty = errors.New("ring buffer is empty")

// NOTE: RingBuf is not thread safe.
type RingBuf[T any] struct {
	data             []T
	capacity, length int
	head, tail       int
}

func New[T any](capacity int) *RingBuf[T] {
	return &RingBuf[T]{
		data:     make([]T, capacity),
		capacity: capacity,
		length:   0,
		head:     0,
		tail:     0,
	}
}

func (b *RingBuf[T]) Dequeue() (*T, error) {
	if b.length <= 0 {
		return nil, ErrEmpty
	}

	val := new(T)
	*val = b.data[b.head]

	b.head = (b.head + 1) % b.capacity
	b.length -= 1
	return val, nil
}

func (b *RingBuf[T]) Enqueue(item T) {
	b.data[b.tail] = item
	b.tail = (b.tail + 1) % b.capacity
	if b.length < b.capacity {
		b.length += 1
	}
}

func (b *RingBuf[T]) ToSlice() []T {
	if b.IsFull() {
		return b.data
	} else if b.tail < b.head {
		return append(b.data[:b.tail], b.data[b.head:]...)
	}
	return b.data[b.head:b.tail]
}

func (b *RingBuf[T]) Len() int {
	return b.length
}

func (b *RingBuf[T]) Clear() {
	b.data = make([]T, b.capacity)
	b.length = 0
	b.head = 0
	b.tail = 0
}

func (b *RingBuf[T]) IsFull() bool {
	return b.length == b.capacity
}
