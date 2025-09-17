package heap

import (
	"iter"
	"sync"
)

type AsyncHeap[T any] struct {
	heap *Heap[T]
	mtx  sync.RWMutex
}

func NewAsyncHeap[T any](less func(a T, b T) bool) *AsyncHeap[T] {
	return &AsyncHeap[T]{
		heap: NewHeap(less),
	}
}

func (h *AsyncHeap[T]) Pop() T {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	return h.heap.Pop()
}

func (h *AsyncHeap[T]) Push(items ...T) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.heap.Push(items...)
}

func (h *AsyncHeap[T]) Peak() T {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	return h.heap.Peak()
}

func (h *AsyncHeap[T]) Empty() bool {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	return h.heap.Empty()
}

func (h *AsyncHeap[T]) Len() int {
	h.mtx.RLock()
	defer h.mtx.RUnlock()
	return h.heap.Len()
}

func (h *AsyncHeap[T]) Clear() {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.heap.Clear()
}

func (h *AsyncHeap[T]) RemoveFunc(del func(a T) bool) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.heap.RemoveFunc(del)
}

func (h *AsyncHeap[T]) PopSeq() iter.Seq[T] {
	return func(yield func(T) bool) {
		for !h.Empty() {
			if !yield(h.Pop()) {
				return
			}
		}
	}
}
