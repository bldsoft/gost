package distlock

import (
	"context"
	"testing"
	"time"

	"github.com/bldsoft/gost/mongo"
	"github.com/stretchr/testify/assert"
)

func TestMongoDistLock_LockUnlock(t *testing.T) {
	db := mongo.NewStorage(mongo.Config{
		Server: "mongodb://root_username:root_password@localhost:27017",
		DbName: "streampool",
	})

	db.Connect()

	lockID := "test-lock"
	ttl := 1 * time.Second

	dl := NewMongoDistLock(db, lockID, ttl)

	assert.True(t, dl.TryLock(), "Expected to acquire lock")

	assert.False(t, dl.TryLock(), "Expected to fail to acquire lock again")

	dl.Unlock()
	assert.True(t, dl.TryLock(), "Expected to acquire lock again after unlock")
}

func TestMongoDistLock_LockExpiration(t *testing.T) {
	db := mongo.NewStorage(mongo.Config{
		Server: "mongodb://root_username:root_password@localhost:27017",
		DbName: "streampool",
	})

	db.Connect()

	lockID := "test-lock-expiration"
	ttl := time.Second

	dl := NewMongoDistLock(db, lockID, ttl)

	assert.True(t, dl.TryLock(), "Expected to acquire lock")

	time.Sleep(ttl + 100*time.Millisecond)

	assert.True(t, dl.TryLock(), "Expected to acquire lock after expiration")
}

func TestMongoDistLock_ConcurrentLocking(t *testing.T) {
	db := mongo.NewStorage(mongo.Config{
		Server: "mongodb://root_username:root_password@localhost:27017",
		DbName: "streampool",
	})
	db.Connect()

	lockID := "test-lock-concurrent"
	ttl := 2 * time.Second

	dl1 := NewMongoDistLock(db, lockID, ttl)
	dl2 := NewMongoDistLock(db, lockID, ttl)

	assert.True(t, dl1.TryLock(), "Expected to acquire lock by first instance")
	assert.False(t, dl2.TryLock(), "Expected second instance to fail to acquire lock")

	dl1.Unlock()
	assert.True(t, dl2.TryLock(), "Expected second instance to acquire lock after first instance unlocks")
}

func TestMongoDistLock_LockRenew(t *testing.T) {
	db := mongo.NewStorage(mongo.Config{
		Server: "mongodb://root_username:root_password@localhost:27017",
		DbName: "streampool",
	})

	db.Connect()

	lockID := "test-lock-renew"
	ttl := time.Second

	dl := NewMongoDistLock(db, lockID, ttl)

	assert.True(t, dl.TryLock(), "Expected to acquire lock")

	time.Sleep(ttl / 2)
	assert.False(t, dl.TryLock(), "Expected to fail to acquire lock while renew is active")

	time.Sleep(ttl)
	assert.True(t, dl.TryLock(), "Expected to acquire lock after ttl expires")
}

func TestMongoDistLock_Quit(t *testing.T) {
	db := mongo.NewStorage(mongo.Config{
		Server: "mongodb://root_username:root_password@localhost:27017",
		DbName: "streampool",
	})
	db.Connect()

	lockID := "test-lock-quit"
	ttl := time.Second
	dl := NewMongoDistLock(db, lockID, ttl)

	assert.True(t, dl.TryLock(), "Expected to acquire lock")

	quitChan := dl.Quit()

	assert.NotNil(t, quitChan, "Expected quit channel to be non-nil")

	dl.Unlock()
	<-quitChan

	select {
	case _, ok := <-quitChan:
		assert.False(t, ok, "Expected quit channel to be closed after unlock")
	default:
	}
}

func TestMongoDistLock_ChangeOwnership(t *testing.T) {
	db := mongo.NewStorage(mongo.Config{
		Server: "mongodb://root_username:root_password@localhost:27017",
		DbName: "streampool",
	})
	db.Connect()

	lockID := "shared-lock-id"
	ttl := time.Second
	dl1 := NewMongoDistLock(db, lockID, ttl)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	dl1.Lock(ctx)
	defer dl1.Unlock()

	go func() {
		time.Sleep(50 * time.Millisecond)
		dl2 := NewMongoDistLock(db, lockID, ttl)
		dl2.Lock(ctx)
		defer dl2.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			assert.Fail(t, "Context timed out before ownership changed")
		case <-dl1.Quit():
			return
		}
	}
}

func TestPanic(t *testing.T) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		assert.True(t, true)
	// 	}
	// }()

	db := mongo.NewStorage(mongo.Config{
		Server: "mongodb://root_username:root_password@localhost:27017",
		DbName: "streampool",
	})
	db.Connect()

	lockID := "shared-lock-id"
	ttl := time.Second
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	l := NewMongoDistLock(db, lockID, ttl)
	l.Lock(ctx)
	defer l.Unlock()

	l2 := NewMongoDistLock(db, lockID, ttl)
	l2.Lock(ctx)
	defer l2.Unlock()

	// <-l.Quit()
	// assert.False(t, ok)

	time.Sleep(200 * time.Millisecond)
	assert.True(t, true)
}

/*
*	select {
*		case <-s.DistMtx.Quit():
*	}
*
* */
