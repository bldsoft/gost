package v2

import (
	"context"
	"sync/atomic"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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

type Watcher struct {
	collection *mongo.Collection
	cancel     context.CancelFunc
	handler    func(fullDocument bson.Raw, opType OperationType)
	isActive   int32
}

func NewWatcher(collection *mongo.Collection) *Watcher {
	return &Watcher{collection: collection}
}

func (w *Watcher) SetHandler(handler func(fullDocument bson.Raw, opType OperationType)) {
	w.handler = handler
}

func (w *Watcher) canStart() bool {
	return atomic.CompareAndSwapInt32(&w.isActive, 0, 1)
}

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
	changeStreamWatcher := NewChangeStreamWatcher()

	go changeStreamWatcher.watch(ctx, w.collection, func(fullDocument bson.Raw, opType OperationType) {
		if updatedTime, ok := fullDocument.Lookup(BsonFieldNameUpdateTime).TimeOK(); ok {
			reserveWatcher.SetLastCheckTime(updatedTime)
		}
		w.handler(fullDocument, opType)
	})
	reserveWatcher.watch(ctx, w.collection, w.handler)
}

func (w *Watcher) Stop() {
	if w == nil {
		return
	}

	w.cancel()
}
