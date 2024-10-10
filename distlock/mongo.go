package distlock

import (
	"context"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/utils"
	lock "github.com/square/mongo-lock"
)

const (
	collName = "distr_mutex"
)

type lockOwner int

const (
	nobody lockOwner = iota
	me
	notme
)

type mongoDistLock struct {
	client     *lock.Client
	recourceID string
	lockID     string
	ttl        time.Duration
	renewStop  func()
	quit       chan struct{}
}

func NewMongoDistLock(db *mongo.Storage, lockID string, ttl time.Duration) DistrMutex {
	col := db.Db.Collection(collName)

	client := lock.NewClient(col)

	ctx := context.Background()
	if err := client.CreateIndexes(ctx); err != nil {
		log.WarnWithFields(log.Fields{"err": err}, "mongoDistLock: failed creating indexes")
	}

	return &mongoDistLock{
		client:     client,
		recourceID: lockID,
		lockID:     utils.RandString(64),
		ttl:        ttl,
	}
}

func (l *mongoDistLock) Lock(ctx context.Context) {
	ticker := time.NewTicker(l.ttl)
	defer ticker.Stop()

	for {
		if l.TryLock() {
			return
		}
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func (l *mongoDistLock) TryLock() bool {
	ctx := context.Background()
	err := l.client.XLock(ctx, l.recourceID, l.lockID, lock.LockDetails{
		TTL:   uint(l.ttl.Seconds()),
		Owner: l.lockID,
	})
	if err != nil {
		log.DebugWithFields(log.Fields{
			"err":        err,
			"resourceID": l.recourceID,
			"lockID":     l.lockID,
		}, "mongoDistLock: failed to lock")
		return false
	}

	ctx, l.renewStop = context.WithCancel(ctx)
	go func() {
		l.quit = make(chan struct{})
		defer close(l.quit)
		l.updateLock(ctx)
	}()

	return true
}

func (l *mongoDistLock) Unlock() {
	l.renewStop()
	_, err := l.client.Unlock(context.Background(), l.lockID)
	if err != nil {
		log.WarnWithFields(log.Fields{
			"err":        err,
			"resourceID": l.recourceID,
			"lockID":     l.lockID,
		}, "mongoDistLock: failed ot unlock")
	}
}

func (l *mongoDistLock) Quit() <-chan struct{} {
	return l.quit
}

func (l *mongoDistLock) updateLock(ctx context.Context) {
	ticker := time.NewTicker(l.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
		_, err := l.client.Renew(ctx, l.lockID, uint(l.ttl.Seconds()))
		if err != nil {
			log.WarnWithFields(log.Fields{"error": err, "lockID": l.lockID}, "mongoDistLock: failed to renew lock status")
			return
		}
	}
}
