package mongo

import (
	"context"
	"reflect"
	"slices"
	"sync"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const updateChanBufferSize = 12500

type WatchedRepositoryOptions struct {
	CacheKeyPrefix string
	WarmUp         bool
}

// watcherEntry wraps a registered watcher so subscriptions can be identified
// by pointer. This avoids relying on interface equality, which panics for
// watchers whose dynamic type is not comparable (e.g. the default
// repository.NewWatcher implementation that holds function fields).
type watcherEntry[T any, U repository.IEntityIDPtr[T]] struct {
	w repository.Watcher[T, U]
}

type watcherActionType uint8

const (
	watcherActionSubscribe watcherActionType = iota
	watcherActionUnsubscribe
)

type watcherAction[T any, U repository.IEntityIDPtr[T]] struct {
	action watcherActionType
	entry  *watcherEntry[T, U]
}

// WatchedRepository is a helper wrapper for Repository, that allows you to monitor changes via Watcher
type WatchedRepository[T any, U repository.IEntityIDPtr[T]] struct {
	Repository[T, U]
	mongoWatcher *Watcher
	handlers     []*watcherEntry[T, U]
	watcherOpsC  chan watcherAction[T, U]
}

func NewWatchedRepository[T any, U repository.IEntityIDPtr[T]](db *Storage, collectionName string, watchers ...repository.Watcher[T, U]) *WatchedRepository[T, U] {
	rep := &WatchedRepository[T, U]{
		Repository:  NewRepository[T, U](db, collectionName),
		watcherOpsC: make(chan watcherAction[T, U]),
	}
	rep.init()
	for _, w := range watchers {
		rep.AddWatcher(w)
	}
	return rep
}

func (r *WatchedRepository[T, U]) AddWatcher(w repository.Watcher[T, U]) (unsubscribe func()) {
	entry := &watcherEntry[T, U]{w: w}
	r.watcherOpsC <- watcherAction[T, U]{
		action: watcherActionSubscribe,
		entry:  entry,
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			r.watcherOpsC <- watcherAction[T, U]{
				action: watcherActionUnsubscribe,
				entry:  entry,
			}
		})
	}
}

// EnablePreAndPostImages turns on changeStreamPreAndPostImages for the underlying collection
// so that change stream delete events carry the document state prior to the change
// (fullDocumentBeforeChange). Without this, delete events arrive with only the document key.
func (r *WatchedRepository[T, U]) EnablePreAndPostImages(ctx context.Context) error {
	coll := r.Repository.Collection()
	cmd := bson.D{
		{Key: "collMod", Value: coll.Name()},
		{Key: "changeStreamPreAndPostImages", Value: bson.M{"enabled": true}},
	}
	return WrapErr(coll.Database().RunCommand(ctx, cmd).Err())
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
			case watcherOp := <-r.watcherOpsC:
				if watcherOp.action == watcherActionSubscribe {
					if err := watcherOp.entry.w.WarmUp(context.Background(), r.Repository); err != nil {
						log.Logger.ErrorWithFields(log.Fields{"err": err, "collection": r.Repository.Name()}, "Failed to warm up")
					} else {
						log.Logger.DebugWithFields(log.Fields{"collection": r.Repository.Name()}, "Warmed up")
					}
					r.handlers = append(r.handlers, watcherOp.entry)
					continue
				}

				if watcherOp.action != watcherActionUnsubscribe {
					log.Logger.ErrorWithFields(log.Fields{"action": watcherOp.action, "collection": r.Repository.Name()}, "Unknown watcher action")
					continue
				}

				r.handlers = slices.DeleteFunc(r.handlers, func(e *watcherEntry[T, U]) bool {
					return e == watcherOp.entry
				})
			case upd := <-updateC:
				for _, handler := range r.handlers {
					handler.w.OnEvent(upd)
				}
			}
		}
	}()
}
