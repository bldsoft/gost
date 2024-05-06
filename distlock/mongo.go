package distlock

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
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
	c        *lock.Client
	lockID   string
	uniqueID string
	ctx      context.Context
	cancel   context.CancelFunc
	ticker   *time.Ticker
	ttl      time.Duration
	quit     chan struct{}
}

func NewMongoDistLock(db *mongo.Storage, lockID string, ttl time.Duration) DistrMutex {
	col := db.Db.Collection(collName)

	uniqueID := make([]byte, 4)
	rand.Read(uniqueID)

	c := lock.NewClient(col)

	ctx, cancel := context.WithCancel(context.Background())
	if err := c.CreateIndexes(ctx); err != nil {
		log.WarnWithFields(log.Fields{"err": err}, "mongoDistLock: failed creating indexes")
	}

	return &mongoDistLock{
		c:      c,
		lockID: lockID,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (l *mongoDistLock) Lock(ctx context.Context) {
	ticker := time.NewTicker(l.ttl)
	defer ticker.Stop()
	if l.TryLock() {
		return
	}

	for {
		select {
		case <-ticker.C:
			if l.TryLock() {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (l *mongoDistLock) TryLock() bool {
	err := l.c.SLock(l.ctx, string(l.uniqueID), l.lockID, lock.LockDetails{
		TTL:   uint(l.ttl.Seconds()),
		Owner: l.uniqueID,
	}, -1)
	if err != nil {
		log.WarnWithFields(log.Fields{"error": err, "lockID": l.lockID}, "Failed to lock memcached mutex")
		return false
	}

	return true
}

func (l *mongoDistLock) Unlock() {
	if l.ticker != nil {
		l.ticker.Stop()
	}
	_, err := l.c.Unlock(l.ctx, l.lockID)
	if err != nil {
		log.WarnWithFields(log.Fields{"err": err, "lockID": l.lockID}, "mongoDistLock: failed ot unlock")
	}
}

func (l *mongoDistLock) Quit() <-chan struct{} {
	return l.quit
}

func (l *mongoDistLock) stop() {
	l.quit <- struct{}{}
	l.ticker.Stop()
}

func (l *mongoDistLock) updateLock() {
	for range l.ticker.C {
		lockOwner := l.getOwner()
		switch lockOwner {
		case me:
			_, err := l.c.Renew(l.ctx, l.lockID, uint(l.ttl*time.Second))
			if err != nil {
				log.WarnWithFields(log.Fields{"error": err, "lockID": l.lockID}, "mongoDistLock: failed to renew lock status")
			}
		case notme:
			l.stop()
			return
		case nobody:
			if !l.TryLock() {
				l.stop()
			}
			return
		}
	}
}

func (l *mongoDistLock) getOwner() lockOwner {
	statuses, err := l.c.Status(l.ctx, lock.Filter{LockId: l.lockID, CreatedAfter: time.Now().Add(-l.ttl * time.Second)})
	if err != nil || len(statuses) == 0 {
		return nobody
	}
	if statuses[0].Owner == l.uniqueID {
		return me
	}
	return notme
}
