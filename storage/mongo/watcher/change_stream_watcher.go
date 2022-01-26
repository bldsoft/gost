package watcher

import (
	"context"
	"time"

	"github.com/bldsoft/gost/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const changeStreamRecoverTime = 30 * time.Second

type changeStreamWatcher struct {
	operationTypes []string
	resumeToken    bson.Raw
	recoverTime    time.Duration
}

func newChangeStreamWatcher() *changeStreamWatcher {
	operationTypes := []string{"update", "replace", "insert"}
	return &changeStreamWatcher{operationTypes: operationTypes, recoverTime: changeStreamRecoverTime}
}

// Watch creates change stream and watch collection. It invokes handler() for each updated.
func (w *changeStreamWatcher) watch(ctx context.Context, collection *mongo.Collection, handler WatchHandler) {
	w.changeStreamWatch(ctx, collection, handler)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(w.recoverTime)
			w.changeStreamWatch(ctx, collection, handler)
		}
	}
}

func (w *changeStreamWatcher) changeStreamWatch(ctx context.Context, collection *mongo.Collection, handler WatchHandler) {
	pipline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "operationType", Value: bson.D{{Key: "$in", Value: w.operationTypes}}}}}},
	}
	opt := options.ChangeStream()
	opt.SetFullDocument(options.UpdateLookup)
	if w.resumeToken != nil {
		opt.SetResumeAfter(w.resumeToken)
	}

	changeStream, err := collection.Watch(ctx, pipline, opt)
	if err != nil {
		log.Warnf("Falied to get change stream: %s", err.Error())
		return
	}
	defer changeStream.Close(ctx)
	log.Debugf("Change stream watcher for \"%s\" started", collection.Name())
	for changeStream.Next(ctx) {
		if handler == nil {
			continue
		}
		operationType := changeStream.Current.Lookup("operationType").StringValue()
		fullDocument := changeStream.Current.Lookup("fullDocument").Document()

		if opType := w.getOpType(fullDocument, operationType); opType != None {
			handler(fullDocument, opType)
			log.Debugf("Change stream watcher detected changes: %s %s", opType, fullDocument)
		}

		w.resumeToken = changeStream.ResumeToken()
	}

	if err = changeStream.Err(); err != nil {
		log.Debugf("Change stream watcher for \"%s\" collection stopped: %s", collection.Name(), err.Error())
	}
}

func (w *changeStreamWatcher) getOpType(fulldocument bson.Raw, opType string) OperationType {
	switch opType {
	case "insert":
		return Insert
	case "update", "replace":
		// TODO: check delete flag
		return Update
	default:
		return None
	}
}
