package mongo

import (
	"context"
	"sync/atomic"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OperationType = string

const (
	None   OperationType = "None"
	Insert OperationType = "Insert"
	Update OperationType = "Update"
	Delete OperationType = "Delete"
)

type WatchHandler = func(fullDocument bson.Raw, opType OperationType)

type IWatcher interface {
	watch(ctx context.Context, collection *mongo.Collection, handler WatchHandler)
}

// Watcher is used for watching collection.
type Watcher struct {
	collection *mongo.Collection
	cancel     context.CancelFunc
	handler    func(fullDocument bson.Raw, opType OperationType)
	isActive   int32
	dbReady    *atomic.Bool
}

// NewWatcher creates new MongoPoolWatcher.
func NewWatcher(collection *mongo.Collection, dbReady *atomic.Bool) *Watcher {
	return &Watcher{collection: collection, dbReady: dbReady}
}

// SetHandler sets handler for watch method. opType is "update",
func (w *Watcher) SetHandler(handler func(fullDocument bson.Raw, opType OperationType)) {
	w.handler = handler
}

func (w *Watcher) canStart() bool {
	return atomic.CompareAndSwapInt32(&w.isActive, 0, 1)
}

// Start starts watching collection.
func (w *Watcher) Start() {
	if w == nil {
		return
	}

	if !w.canStart() {
		return
	}

	go w.watch()
}

func (w *Watcher) getNewContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	return ctx
}

func (w *Watcher) watch() {
	defer func() {
		w.isActive = 0
	}()
	ctx := w.getNewContext()

	reserveWatcher := newReserveWatcher()
	// changeStreamWatcher := NewChangeStreamWatcher()

	go func() {
		for !w.dbReady.Load() {
		}
		// changeStreamWatcher.watch(ctx, w.collection, func(fullDocument bson.Raw, opType OperationType) {
		// 	if updatedTime, ok := fullDocument.Lookup(BsonFieldNameUpdateTime).TimeOK(); ok {
		// 		reserveWatcher.SetLastCheckTime(updatedTime)
		// 	}
		// 	w.handler(fullDocument, opType)
		// })

		reserveWatcher.watch(ctx, w.collection, w.handler)
	}()

}

// Stop stops watching collection.
func (w *Watcher) Stop() {
	if w == nil {
		return
	}

	w.cancel()
}
