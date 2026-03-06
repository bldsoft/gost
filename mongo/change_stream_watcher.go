package mongo

import (
	"context"
	"errors"
	"time"

	"github.com/bldsoft/gost/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoErrCodeResumeTokenLost        = 286
	mongoErrCodeChangeStreamFatalError = 280
)

const changeStreamRecoverTime = 30 * time.Second

const (
	changeStreamInsertOp  = "insert"
	changeStreamUpdateOp  = "update"
	changeStreamReplaceOp = "replace"
	changeStreamDeleteOp  = "delete"
)

type changeStreamWatcher struct {
	operationTypes []string
	resumeToken    bson.Raw
	recoverTime    time.Duration
}

func NewChangeStreamWatcher(operations ...OperationType) *changeStreamWatcher {
	if len(operations) == 0 {
		operations = []OperationType{Update, Insert, Delete}
	}
	var operationTypes []string
	for _, op := range operations {
		switch op {
		case Insert:
			operationTypes = append(operationTypes, changeStreamInsertOp)
		case Update:
			operationTypes = append(operationTypes, changeStreamUpdateOp, changeStreamReplaceOp)
		case Delete:
			operationTypes = append(operationTypes, changeStreamDeleteOp)
		}
	}
	return &changeStreamWatcher{
		operationTypes: operationTypes,
		recoverTime:    changeStreamRecoverTime,
	}
}

func (w *changeStreamWatcher) Watch(ctx context.Context, collection *mongo.Collection, handler WatchHandler) {
	go w.watch(ctx, collection, handler)
}

// Watch creates change stream and watch collection. It invokes handler() for each updated.
func (w *changeStreamWatcher) watch(ctx context.Context, collection *mongo.Collection, handler WatchHandler) {
	ctx = context.WithValue(ctx, log.LoggerCtxKey, log.FromContext(ctx).WithFields(log.Fields{"collection": collection.Name()}))
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
	opt.SetFullDocumentBeforeChange(options.WhenAvailable)
	if w.resumeToken != nil {
		opt.SetResumeAfter(w.resumeToken)
	}

	changeStream, err := collection.Watch(ctx, pipline, opt)
	if err != nil {
		if needResumeTokenReset(err) {
			w.resumeToken = nil
		}
		log.FromContext(ctx).Warnf("Falied to get change stream: %s", err.Error())
		return
	}
	defer changeStream.Close(ctx)
	log.FromContext(ctx).Debugf("Change stream watcher for \"%s\" started", collection.Name())
	for changeStream.Next(ctx) {
		if handler == nil {
			continue
		}
		operationType := changeStream.Current.Lookup("operationType").StringValue()
		var fullDocument bson.Raw
		if operationType == changeStreamDeleteOp {
			// for delete operation only _id is returned
			if bc := changeStream.Current.Lookup("fullDocumentBeforeChange"); bc.Type == bson.TypeNull && len(bc.Value) == 0 {
				fullDocument = changeStream.Current.Lookup("documentKey").Document()
			} else {
				fullDocument = bc.Document()
			}
		} else {
			fullDocument = changeStream.Current.Lookup("fullDocument").Document()
		}

		if opType := w.getOpType(fullDocument, operationType); opType != None {
			handler(fullDocument, opType)
			log.FromContext(ctx).Debugf("Change stream watcher detected changes: %s %s", opType, fullDocument)
		}

		w.resumeToken = changeStream.ResumeToken()
	}

	if err = changeStream.Err(); err != nil {
		log.FromContext(ctx).Debugf("Change stream watcher for \"%s\" collection stopped: %s", collection.Name(), err.Error())
	}
}

func (w *changeStreamWatcher) getOpType(fulldocument bson.Raw, opType string) OperationType {
	switch opType {
	case changeStreamInsertOp:
		return Insert
	case changeStreamUpdateOp, changeStreamReplaceOp:
		return Update
	case changeStreamDeleteOp:
		return Delete
	default:
		return None
	}
}

func needResumeTokenReset(err error) bool {
	if se := mongo.ServerError(nil); errors.As(err, &se) {
		return se.HasErrorCode(mongoErrCodeResumeTokenLost) ||
			se.HasErrorCode(mongoErrCodeChangeStreamFatalError)
	}
	return false
}
