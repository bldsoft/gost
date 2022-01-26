package watcher

import (
	"context"
	"sync/atomic"

	gost "github.com/bldsoft/gost/storage/mongo"
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

// MongoWatcher is used for watching collection.
type MongoWatcher struct {
	collection *mongo.Collection
	cancel     context.CancelFunc
	handler    func(fullDocument bson.Raw, opType OperationType)
	isActive   int32
}

// NewMongoWatcher creates new MongoPoolWatcher.
func NewMongoWatcher(collection *mongo.Collection) *MongoWatcher {
	return &MongoWatcher{collection: collection}
}

// SetHandler sets handler for watch method. opType is "update",
func (w *MongoWatcher) SetHandler(handler func(fullDocument bson.Raw, opType OperationType)) {
	w.handler = handler
}

func (w *MongoWatcher) canStart() bool {
	return atomic.CompareAndSwapInt32(&w.isActive, 0, 1)
}

// Start starts watching collection.
func (w *MongoWatcher) Start() {
	if w == nil {
		return
	}

	if !w.canStart() {
		return
	}

	go w.watch()
}

func (w *MongoWatcher) getNewContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	return ctx
}

func (w *MongoWatcher) watch() {
	defer func() {
		w.isActive = 0
	}()
	ctx := w.getNewContext()

	reserveWatcher := newReserveWatcher()
	changeStreamWatcher := newChangeStreamWatcher()

	go changeStreamWatcher.watch(ctx, w.collection, func(fullDocument bson.Raw, opType OperationType) {
		if updatedTime, ok := fullDocument.Lookup(gost.BsonFieldNameUpdateTime).TimeOK(); ok {
			reserveWatcher.SetLastCheckTime(updatedTime)
		}
		w.handler(fullDocument, opType)
	})
	reserveWatcher.watch(ctx, w.collection, w.handler)

}

// Stop stops watching collection.
func (w *MongoWatcher) Stop() {
	if w == nil {
		return
	}

	w.cancel()
}
