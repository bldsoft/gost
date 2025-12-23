package queue

import "context"

type Item[T any] interface {
	Value() T
	Ack(ctx context.Context) error
	Nack(ctx context.Context) error
}

type Queue[T any] interface {
	Enqueue(ctx context.Context, entity ...T) error
	Dequeue(ctx context.Context) (Item[T], error)
}

type BatchQueue[T any] interface {
	Queue[T]
	DequeueMany(ctx context.Context, n int) ([]Item[T], error)
	AckMany(ctx context.Context, items ...Item[T]) error
	NackMany(ctx context.Context, items ...Item[T]) error
}
