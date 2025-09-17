package heap

import (
	"container/heap"
	"iter"
)

type heapItem[T any] struct {
	value T
	seq   uint64
}

type innerHeap[T any] struct {
	slice []heapItem[T]
	less  func(T, T) bool
	seq   uint64
}

func (h *innerHeap[T]) Len() int { return len(h.slice) }
func (h *innerHeap[T]) Less(i, j int) bool {
	a, b := h.slice[i], h.slice[j]
	if h.less(a.value, b.value) {
		return true
	}
	if h.less(b.value, a.value) {
		return false
	}
	return a.seq < b.seq
}
func (h *innerHeap[T]) Swap(i, j int) { h.slice[i], h.slice[j] = h.slice[j], h.slice[i] }

func (h *innerHeap[T]) Pop() any {
	old := h.slice
	n := len(old)
	x := old[n-1]
	h.slice = old[0 : n-1]
	return x.value
}

func (h *innerHeap[T]) Push(x any) {
	value := x.(T)
	h.seq++
	h.slice = append(h.slice, heapItem[T]{
		value: value,
		seq:   h.seq,
	})
}

func (h *innerHeap[T]) Peak() T {
	return h.slice[0].value
}

func (h *innerHeap[T]) Empty() bool {
	return len(h.slice) == 0
}

func (h *innerHeap[T]) Clear() {
	h.slice = h.slice[:0]
	h.seq = 0
}

type Heap[T any] struct {
	innerHeap[T]
}

func NewHeap[T any](less func(a T, b T) bool) *Heap[T] {
	i := innerHeap[T]{
		slice: make([]heapItem[T], 0, 128),
		less:  less,
		seq:   0,
	}
	heap.Init(&i)
	return &Heap[T]{
		innerHeap: i,
	}
}

func (h *Heap[T]) Pop() T {
	return heap.Pop(&h.innerHeap).(T)
}

func (h *Heap[T]) Push(items ...T) {
	for _, item := range items {
		heap.Push(&h.innerHeap, item)
	}
}

func (h *Heap[T]) Peak() T {
	return h.innerHeap.Peak()
}

func (h *Heap[T]) Empty() bool {
	return h.innerHeap.Empty()
}

func (h *Heap[T]) Len() int {
	return h.innerHeap.Len()
}

func (h *Heap[T]) Clear() {
	h.innerHeap.Clear()
}

func (h *Heap[T]) RemoveFunc(del func(a T) bool) {
	for i := 0; i < h.innerHeap.Len(); i++ {
		if del(h.innerHeap.slice[i].value) {
			heap.Remove(&h.innerHeap, i)
		}
	}
}

func (h *Heap[T]) PopSeq() iter.Seq[T] {
	return func(yield func(T) bool) {
		for !h.Empty() {
			if !yield(h.Pop()) {
				return
			}
		}
	}
}
