package mongo

import (
	"context"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
)

const updateChanBufferSize = 12500

type ChangeHandler[T any, U repository.IEntityIDPtr[T]] interface {
	OnWarmUp(entities []U)
	OnChange(upd *UpdateEvent[T, U])
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
type WatchedRepository[T any, U repository.IEntityIDPtr[T], FT any] struct {
	*Repository[T, U, FT]
	Watcher    *Watcher
	handler    ChangeHandler[T, U]
	needWarmUp bool
}

func NewWatchedRepository[T any, U repository.IEntityIDPtr[T], FT any](db *Storage, collectionName string, handler ChangeHandler[T, U], warmUp bool) *WatchedRepository[T, U, FT] {
	rep := &WatchedRepository[T, U, FT]{
		Repository: NewRepository[T, U, FT](db, collectionName),
		handler:    handler,
		needWarmUp: warmUp,
	}
	rep.init()
	return rep
}

func (r *WatchedRepository[T, U, FT]) init() {
	updateC := make(chan *UpdateEvent[T, U], updateChanBufferSize)

	r.Watcher = NewWatcher(r.Repository.Collection())
	r.Watcher.SetHandler(func(fullDocument bson.Raw, opType OperationType) {
		var e T
		if err := bson.Unmarshal(fullDocument, &e); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err}, "WatchedRepository: failed to update %T")
		}
		updateC <- &UpdateEvent[T, U]{
			Entity: &e,
			OpType: opType,
		}
	})
	r.Watcher.Start()

	if r.needWarmUp {
		if err := r.warmUp(context.Background()); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err, "collection": r.Repository.collectionName}, "Failed to warm up")
		} else {
			log.Logger.DebugWithFields(log.Fields{"collection": r.Repository.collectionName}, "Warmed up")
		}
	}

	go func() {
		for upd := range updateC {
			r.handler.OnChange(upd)
		}
	}()
}

func (r *WatchedRepository[T, U, FT]) warmUp(ctx context.Context) error {
	entities, err := r.Repository.GetAll(context.Background())
	if err != nil {
		return err
	}
	r.handler.OnWarmUp(entities)
	return nil
}
