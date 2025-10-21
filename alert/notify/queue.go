package notify

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bldsoft/gost/utils"
	"github.com/bldsoft/gost/utils/ringbuf"
)

type MemoryQueue struct {
	ring *ringbuf.RingBuf[RetriedNotification]
	mtx  sync.RWMutex
}

func NewMemoryQueue(capacity int) *MemoryQueue {
	return &MemoryQueue{
		ring: ringbuf.New[RetriedNotification](capacity),
	}
}

func (q *MemoryQueue) Enqueue(ctx context.Context, n RetriedNotification) error {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	if pushed := q.ring.Push(n); pushed == 0 {
		return errors.New("queue is full")
	}
	return nil
}

func (q *MemoryQueue) Dequeue(ctx context.Context) (id string, _ *RetriedNotification, err error) {
	q.mtx.Lock()
	defer q.mtx.Unlock()
	n, ok := q.ring.Top()
	if !ok || time.Now().Before(n.RetryAt) {
		return "", nil, utils.ErrObjectNotFound
	}
	_, _ = q.ring.Pull()
	return "", &n, nil
}

func (q *MemoryQueue) MarkDone(ctx context.Context, id string) error {
	return nil
}

func (q *MemoryQueue) Requeue(ctx context.Context, id string, n RetriedNotification) error {
	return q.Enqueue(ctx, n)
}
