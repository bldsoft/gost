package mongo

import (
	"context"
	"reflect"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
)

const updateChanBufferSize = 12500

type ChangeHandler[T any, U repository.IEntityIDPtr[T]] interface {
	OnChange(upd *UpdateEvent[T, U])
}

type WarmUper[T any, U repository.IEntityIDPtr[T]] interface {
	WarmUp(ctx context.Context, rep Repository[T, U]) error
}

type UpdateEvent[T any, U repository.IEntityIDPtr[T]] struct {
	Entity U
	OpType OperationType
}

type WatchedRepositoryOptions struct {
	CacheKeyPrefix string
	WarmUp         bool
}

// WatchedRepository is a helper wrapper for Repository, that allows you to monitor changes via Watcher
type WatchedRepository[T any, U repository.IEntityIDPtr[T]] struct {
	Repository[T, U]
	Watcher *Watcher
	handler ChangeHandler[T, U]
}

func NewWatchedRepository[T any, U repository.IEntityIDPtr[T]](db *Storage, collectionName string, handler ChangeHandler[T, U]) *WatchedRepository[T, U] {
	rep := &WatchedRepository[T, U]{
		Repository: NewRepository[T, U](db, collectionName),
		handler:    handler,
	}
	rep.init()
	return rep
}

func (r *WatchedRepository[T, U]) init() {
	updateC := make(chan *UpdateEvent[T, U], updateChanBufferSize)

	r.Watcher = NewWatcher(r.Repository.Collection())
	r.Watcher.SetHandler(func(fullDocument bson.Raw, opType OperationType) {
		var e T
		if err := bson.Unmarshal(fullDocument, &e); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err, "type": reflect.TypeOf(e).String()}, "WatchedRepository: failed to update")
		}
		updateC <- &UpdateEvent[T, U]{
			Entity: &e,
			OpType: opType,
		}
	})
	r.Watcher.Start()

	if wu, ok := r.handler.(WarmUper[T, U]); ok {
		if err := wu.WarmUp(context.Background(), r.Repository); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err, "collection": r.Repository.Name()}, "Failed to warm up")
		} else {
			log.Logger.DebugWithFields(log.Fields{"collection": r.Repository.Name()}, "Warmed up")
		}
	}

	go func() {
		for upd := range updateC {
			r.handler.OnChange(upd)
		}
	}()
}
