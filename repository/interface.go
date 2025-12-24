package repository

import (
	"context"
)

type Repository[T any, U IEntityIDPtr[T]] interface {
	FindOne(ctx context.Context, filter interface{}, opts ...*QueryOptions) (U, error)
	FindByID(ctx context.Context, id interface{}, options ...*QueryOptions) (U, error)
	FindByStringIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*QueryOptions) ([]U, error)
	FindByIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*QueryOptions) ([]U, error)
	Find(ctx context.Context, filter interface{}, opt ...*QueryOptions) ([]U, error)
	Count(ctx context.Context, filter interface{}, opt ...*QueryOptions) (int64, error)
	GetAll(ctx context.Context, options ...*QueryOptions) ([]U, error)

	Insert(ctx context.Context, entity U) error
	InsertMany(ctx context.Context, entities []U) error
	Update(ctx context.Context, entity U, options ...*QueryOptions) error
	UpdateMany(ctx context.Context, entities []U) error
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, options ...*QueryOptions) error
	UpdateAndGetByID(ctx context.Context, updateEntity U, returnNewDocument bool, queryOpt ...*QueryOptions) (U, error)
	InsertOrReplace(ctx context.Context, entity U) (inserted bool, _ error)
	UpsertOne(ctx context.Context, filter interface{}, update U) error
	Delete(ctx context.Context, id interface{}, options ...*QueryOptions) error
	DeleteMany(ctx context.Context, filter interface{}, options ...*QueryOptions) error
}

//go:generate go run github.com/abice/go-enum -f=$GOFILE

// ENUM(create, update, delete)
type EventType string

type Event[T any, U IEntityIDPtr[T]] struct {
	Entity U
	Type   EventType
}

type Watcher[T any, U IEntityIDPtr[T]] interface {
	WarmUp(ctx context.Context, rep Repository[T, U]) error
	OnEvent(event Event[T, U])
}

type WatchedRepository[T any, U IEntityIDPtr[T]] interface {
	Repository[T, U]
	AddWatcher(w Watcher[T, U])
}

func NewWatcher[T any, U IEntityIDPtr[T]](
	warmUp func(ctx context.Context, rep Repository[T, U]) error,
	onEvent func(event Event[T, U]),
) Watcher[T, U] {
	return watcher[T, U]{
		warmUp:  warmUp,
		onEvent: onEvent,
	}
}

type watcher[T any, U IEntityIDPtr[T]] struct {
	warmUp  func(ctx context.Context, rep Repository[T, U]) error
	onEvent func(event Event[T, U])
}

func (w watcher[T, U]) WarmUp(ctx context.Context, rep Repository[T, U]) error {
	if w.warmUp != nil {
		return w.warmUp(ctx, rep)
	}
	return nil
}

func (w watcher[T, U]) OnEvent(event Event[T, U]) {
	if w.onEvent != nil {
		w.onEvent(event)
	}
}
