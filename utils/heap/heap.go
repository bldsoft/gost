package heap

import (
	"container/heap"
	"slices"
)

type innerHeap[T any] struct {
	slice []T
	less  func(T, T) bool
}

func (h *innerHeap[T]) Len() int           { return len(h.slice) }
func (h *innerHeap[T]) Less(i, j int) bool { return h.less(h.slice[i], h.slice[j]) }
func (h *innerHeap[T]) Swap(i, j int)      { h.slice[i], h.slice[j] = h.slice[j], h.slice[i] }

func (h *innerHeap[T]) Pop() any {
	old := h.slice
	n := len(old)
	x := old[n-1]
	h.slice = old[0 : n-1]
	return x
}

func (h *innerHeap[T]) Push(x any) {
	h.slice = append(h.slice, x.(T))
}

func (h *innerHeap[T]) Peak() T {
	return h.slice[0]
}

func (h *innerHeap[T]) Empty() bool {
	return len(h.slice) == 0
}

func (h *innerHeap[T]) Clear() {
	h.slice = h.slice[:0]
}

type Heap[T any] struct {
	innerHeap[T]
}

func NewHeap[T any](less func(a T, b T) bool) *Heap[T] {
	i := innerHeap[T]{
		slice: make([]T, 0, 100),
		less:  less,
	}
	heap.Init(&i)
	return &Heap[T]{
		innerHeap: i,
	}
}

func (h *Heap[T]) Pop() T {
	return heap.Pop(&h.innerHeap).(T)
}

func (h *Heap[T]) Top() T {
	return h.innerHeap.Peak()
}

func (h *Heap[T]) RemoveFirstFunc(f func(T) bool) {
	idx := slices.IndexFunc(h.innerHeap.slice, f)
	if idx == -1 {
		return
	}
	_ = heap.Remove(&h.innerHeap, idx)
}

func (h *Heap[T]) Push(items ...T) {
	for _, item := range items {
		heap.Push(&h.innerHeap, item)
	}
}
