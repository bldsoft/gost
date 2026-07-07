package timedqueue

import (
	"context"
	"iter"
	"sync"
	"time"

	"github.com/bldsoft/gost/utils/heap"
)

type item[T any] struct {
	value T
	next  time.Time
}

type TimedQueue[T any] struct {
	heap   *heap.Heap[item[T]]
	mtx    sync.RWMutex
	pushed chan struct{}
	closed bool
}

func New[T any]() *TimedQueue[T] {
	return &TimedQueue[T]{
		heap: heap.NewHeap(func(a, b item[T]) bool {
			return a.next.Before(b.next)
		}),
		pushed: make(chan struct{}, 1),
	}
}

func (q *TimedQueue[T]) notify() {
	select {
	case q.pushed <- struct{}{}:
	default:
	}
}

func (q *TimedQueue[T]) Push(value T, next time.Time) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	if q.closed {
		return
	}
	q.heap.Push(item[T]{value, next})
	q.notify()
}

func (q *TimedQueue[T]) popAndWait(ctx context.Context) *item[T] {
	for {
		v := q.topWait(ctx)
		if v == nil {
			return nil
		}

		select {
		case <-time.After(time.Until(v.next)):
			res := func() *item[T] {
				q.mtx.Lock()
				defer q.mtx.Unlock()
				if q.heap.Empty() {
					return nil
				}
				res := q.heap.Top()
				if !res.next.After(v.next) {
					_ = q.heap.Pop()
					return &res
				}
				return nil
			}()
			if res != nil {
				return res
			}
		case <-q.pushed:
		case <-ctx.Done():
			return nil
		}
	}
}

func (q *TimedQueue[T]) RemoveFirstFunc(f func(value T) bool) (found bool) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	found = q.heap.RemoveFirstFunc(func(item item[T]) bool {
		return f(item.value)
	})
	q.notify()
	return found
}

func (q *TimedQueue[T]) topWait(ctx context.Context) *item[T] {
	item, closed := q.top()
	for item == nil {
		if closed {
			return nil
		}
		select {
		case <-q.pushed:
		case <-ctx.Done():
			return nil
		}
		item, closed = q.top()
	}
	return item
}

func (q *TimedQueue[T]) top() (item *item[T], closed bool) {
	q.mtx.RLock()
	defer q.mtx.RUnlock()
	if q.heap.Empty() {
		return nil, q.closed
	}
	res := q.heap.Top()
	return &res, false
}

func (q *TimedQueue[T]) SyncSeq(ctx context.Context) iter.Seq[T] {
	return func(yield func(T) bool) {
		for item := q.popAndWait(ctx); item != nil; item = q.popAndWait(ctx) {
			if !yield(item.value) {
				return
			}
		}
	}
}

func (q *TimedQueue[T]) SyncSeq2(ctx context.Context) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		for item := range q.SyncSeq(ctx) {
			if !yield(i, item) {
				return
			}
			i++
		}
	}
}

func (q *TimedQueue[T]) Close() {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	if q.closed {
		return
	}
	q.closed = true
	q.notify()
}
