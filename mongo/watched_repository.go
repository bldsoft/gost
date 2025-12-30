package mongo

import (
	"context"
	"reflect"
	"slices"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
)

const updateChanBufferSize = 12500

type WatchedRepositoryOptions struct {
	CacheKeyPrefix string
	WarmUp         bool
}

// WatchedRepository is a helper wrapper for Repository, that allows you to monitor changes via Watcher
type WatchedRepository[T any, U repository.IEntityIDPtr[T]] struct {
	Repository[T, U]
	mongoWatcher   *Watcher
	handlers       []repository.Watcher[T, U]
	handlerC       chan repository.Watcher[T, U]
	deleteHandlerC chan repository.Watcher[T, U]
}

func NewWatchedRepository[T any, U repository.IEntityIDPtr[T]](db *Storage, collectionName string, watchers ...repository.Watcher[T, U]) *WatchedRepository[T, U] {
	rep := &WatchedRepository[T, U]{
		Repository:     NewRepository[T, U](db, collectionName),
		handlerC:       make(chan repository.Watcher[T, U]),
		deleteHandlerC: make(chan repository.Watcher[T, U]),
	}
	rep.init()
	for _, w := range watchers {
		rep.AddWatcher(w)
	}
	return rep
}

func (r *WatchedRepository[T, U]) AddWatcher(w repository.Watcher[T, U]) (unsubscribe func()) {
	r.handlerC <- w
	return func() {
		r.deleteHandlerC <- w
	}
}

func (r *WatchedRepository[T, U]) init() {
	updateC := make(chan repository.Event[T, U], updateChanBufferSize)

	convertEventType := map[OperationType]repository.EventType{
		Insert: repository.EventTypeCreate,
		Update: repository.EventTypeUpdate,
		Delete: repository.EventTypeDelete,
	}

	r.mongoWatcher = NewWatcher(r.Repository.Collection())
	r.mongoWatcher.SetHandler(func(fullDocument bson.Raw, opType OperationType) {
		var e T
		if err := bson.Unmarshal(fullDocument, &e); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err, "type": reflect.TypeOf(e).String()}, "WatchedRepository: failed to update")
		}

		updateC <- repository.Event[T, U]{
			Entity: &e,
			Type:   convertEventType[opType],
		}
	})
	r.mongoWatcher.Start()

	go func() {
		for {
			select {
			case handler := <-r.handlerC:
				if err := handler.WarmUp(context.Background(), r.Repository); err != nil {
					log.Logger.ErrorWithFields(log.Fields{"err": err, "collection": r.Repository.Name()}, "Failed to warm up")
				} else {
					log.Logger.DebugWithFields(log.Fields{"collection": r.Repository.Name()}, "Warmed up")
				}
				r.handlers = append(r.handlers, handler)
			case handler := <-r.deleteHandlerC:
				r.handlers = slices.DeleteFunc(r.handlers, func(w repository.Watcher[T, U]) bool {
					return w == handler
				})
			case upd := <-updateC:
				for _, handler := range r.handlers {
					handler.OnEvent(upd)
				}
			}
		}
	}()
}
