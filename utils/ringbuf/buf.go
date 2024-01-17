package ringbuf

import "errors"

// NOTE: RingBuf is not thread safe.
type RingBuf[T any] struct {
	data              []T
	readIdx, writeIdx int
	isFull            bool
	overwrite         bool
}

func New[T any](capacity int) *RingBuf[T] {
	if capacity <= 0 {
		panic(errors.New("capacity must be positive"))
	}
	return &RingBuf[T]{
		data: make([]T, capacity),
	}
}

func (b *RingBuf[T]) WithOverwrite(val bool) *RingBuf[T] {
	b.overwrite = val
	return b
}

func (b *RingBuf[T]) Overwrite() bool {
	return b.overwrite
}

func (b RingBuf[T]) Empty() bool {
	return !b.isFull && b.readIdx == b.writeIdx
}

func (b RingBuf[T]) Full() bool {
	return b.isFull
}

func (b RingBuf[T]) Cap() int {
	return len(b.data)
}

func (b RingBuf[T]) Len() int {
	diff := b.writeIdx - b.readIdx
	switch {
	case diff > 0:
		return diff
	case diff < 0:
		return b.Cap() + diff
	case b.isFull:
		return b.Cap()
	default:
		return 0
	}
}

func (b *RingBuf[T]) Push(items ...T) (n int) {
	for _, item := range items {
		if !b.push(item) {
			return n
		}
		n++
	}
	return n
}

func (b *RingBuf[T]) push(item T) (ok bool) {
	if b.isFull {
		if b.overwrite {
			b.readIdx = (b.readIdx + 1) % b.Cap()
		} else {
			return false
		}
	}

	b.data[b.writeIdx] = item
	b.writeIdx = (b.writeIdx + 1) % b.Cap()

	b.isFull = b.writeIdx == b.readIdx

	return true
}

func (b *RingBuf[T]) Get() (T, bool) {
	if b.Empty() {
		var zero T
		return zero, false
	}
	return b.data[b.readIdx], true
}

func (b *RingBuf[T]) Pull() (T, bool) {
	val, ok := b.Get()
	if ok {
		b.readIdx = (b.readIdx + 1) % b.Cap()
		b.isFull = false
	}
	return val, ok
}

func (b *RingBuf[T]) Clear() {
	b.readIdx = 0
	b.writeIdx = 0
	b.isFull = false
}

func (b *RingBuf[T]) Copy(dst []T) int {
	if b.Empty() {
		return 0
	}
	if b.readIdx < b.writeIdx {
		return copy(dst, b.data[b.readIdx:b.writeIdx])
	}
	n := copy(dst, b.data[b.readIdx:])
	n += copy(dst[n:], b.data[0:b.writeIdx])
	return n
}

func (b *RingBuf[T]) ToSlice() []T {
	res := make([]T, b.Cap())
	b.Copy(res)
	return res
}
