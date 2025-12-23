package queue

import (
	"context"
	"errors"
	"time"

	"github.com/bldsoft/gost/utils/ringbuf"
)

type cachedItem[T any] struct {
	q         *CachedQueue[T]
	item      *Item[T]
	expiresAt time.Time
}

type CachedQueueConfig struct {
	Capacity       int
	ProcessTimeout time.Duration
}

type CachedQueue[T any] struct {
	cfg   CachedQueueConfig
	queue BatchQueue[T]
	buf   *ringbuf.RingBuf[*cachedItem[T]]
}

func NewCachedQueue[T any](q BatchQueue[T], cfg CachedQueueConfig) *CachedQueue[T] {
	return &CachedQueue[T]{
		cfg:   cfg,
		queue: q,
	}
}

func (q *CachedQueue[T]) Enqueue(ctx context.Context, entity ...T) error {
	return q.queue.Enqueue(ctx, entity...)
}

func (q *CachedQueue[T]) Dequeue(ctx context.Context) (Item[T], error) {
	now := time.Now()
	for !q.buf.Empty() {
		item, _ := q.buf.Top()
		if item.expiresAt.After(now) {
			break
		}
		item, _ = q.buf.Pull()
	}

	if !q.buf.Empty() {
		// q.buf.Pull()
		return nil, errors.New("queue is full")
	}

	items, err := q.queue.DequeueMany(ctx, q.buf.Cap())
	if err != nil || len(items) == 0 {
		return nil, err
	}

	for _, item := range items {
		q.buf.Push(&cachedItem[T]{
			q:         q,
			item:      &item,
			expiresAt: now.Add(q.cfg.ProcessTimeout),
		})
	}
	return q.Dequeue(ctx)
}
