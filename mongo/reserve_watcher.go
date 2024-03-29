package mongo

import (
	"context"
	"sync"
	"time"

	"github.com/bldsoft/gost/config/feature"
	"github.com/bldsoft/gost/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ReserveWatcherInterval is used for changing update interval when mongo change stream is off. Set it before start watching
// If ReserveWatcherInterval is not set, update interval is 5 min by default.
var ReserveWatcherInterval *feature.Duration

const defaultDuration = time.Minute * 5

type reserveWatcher struct {
	lastCheck time.Time
	mtx       sync.Mutex
}

func newReserveWatcher() *reserveWatcher {
	return &reserveWatcher{lastCheck: time.Now()}
}

func (w *reserveWatcher) getReserveWatcherInterval() time.Duration {
	if ReserveWatcherInterval != nil {
		return ReserveWatcherInterval.Get()
	}
	return defaultDuration
}

func (w *reserveWatcher) SetLastCheckTime(t time.Time) {
	w.mtx.Lock()
	defer w.mtx.Unlock()
	if w.lastCheck.Before(t) {
		w.lastCheck = t
	}
}

func (w *reserveWatcher) watch(ctx context.Context, collection *mongo.Collection, handler WatchHandler) {
	var ticker *time.Ticker
	if readInterval := w.getReserveWatcherInterval(); readInterval > 0 {
		ticker = time.NewTicker(readInterval)
		log.Debugf("Reserve watcher for \"%s\" started", collection.Name())
	} else {
		//create stopped ticker
		ticker = time.NewTicker(time.Second)
		ticker.Stop()
		log.Debugf("Reserve watcher for \"%s\" is paused", collection.Name())
	}

	if ReserveWatcherInterval != nil {
		ReserveWatcherInterval.AddOnChangeHandler(func(interval time.Duration) {
			if interval > 0 {
				ticker.Reset(interval)
				log.Debugf("Reserve watcher interval for \"%s\" is changed to %v", collection.Name(), interval)
			} else {
				ticker.Stop()
				log.Debugf("Reserve watcher for \"%s\" is paused", collection.Name())
			}
		})
	}

	for {
		select {
		case <-ticker.C:
			checkTime := time.Now()
			w.watchCollection(collection, handler)
			w.SetLastCheckTime(checkTime)
		case <-ctx.Done():
			ticker.Stop()
			log.Debugf("Reserve watcher for \"%s\" collection stopped", collection.Name())
			return
		}
	}
}

func (w *reserveWatcher) watchCollection(collection *mongo.Collection, handler WatchHandler) {
	ctx := context.TODO()
	cursor, err := collection.Find(ctx, bson.M{BsonFieldNameUpdateTime: bson.M{"$gte": w.lastCheck}})
	if err != nil {
		log.Errorf("%s reserve watcher falied to open cursor %s", collection.Name(), err.Error())
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var item bson.Raw
		if err = cursor.Decode(&item); err != nil {
			log.Errorf("Falied to decode item %s", err.Error())
			continue
		}
		if opType := w.getOpType(item); opType != None {
			handler(item, opType)
			log.Debugf("Reserve watcher detected changes: %s %s", opType, item)
		}
	}
}

func (w *reserveWatcher) getOpType(item bson.Raw) OperationType {
	var createdTime, updatedTime time.Time
	var ok bool
	createdTime, ok = item.Lookup(BsonFieldNameCreateTime).TimeOK()
	if !ok {
		return None
	}
	updatedTime, ok = item.Lookup(BsonFieldNameUpdateTime).TimeOK()
	if !ok {
		return None
	}

	//TODO: check delete flag

	if createdTime == updatedTime {
		return Insert
	}
	return Update
}
