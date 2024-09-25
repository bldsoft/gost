//go:build integration_test

package distlock

import (
	"context"
	"testing"
	"time"

	"github.com/bldsoft/gost/mongo"
	"github.com/stretchr/testify/assert"
)

var (
	cfg = mongo.Config{
		Server: "mongodb://root_username:root_password@localhost:27017",
		DbName: "streampool",
	}

	db *mongo.Storage
)

func init() {
	db = mongo.NewStorage(cfg)
	db.Connect()
}

func TestMongoDistLock(t *testing.T) {
	testMongoDistLock_LockUnlock(t)
	testMongoDistLock_LockExpiration(t)
	testMongoDistLock_ConcurrentLocking(t)
	testMongoDistLock_LockRenew(t)
	testMongoDistLock_Quit(t)
}

func testMongoDistLock_LockUnlock(t *testing.T) {
	lockID := "test-lock"
	ttl := 1 * time.Second

	dl := NewMongoDistLock(db, lockID, ttl)

	assert.True(t, dl.TryLock(), "Expected to acquire lock")

	assert.False(t, dl.TryLock(), "Expected to fail to acquire lock again")

	dl.Unlock()
	assert.True(t, dl.TryLock(), "Expected to acquire lock again after unlock")
}

func testMongoDistLock_LockExpiration(t *testing.T) {
	lockID := "test-lock-expiration"
	ttl := time.Second

	dl := NewMongoDistLock(db, lockID, ttl)

	assert.True(t, dl.TryLock(), "Expected to acquire lock")

	time.Sleep(ttl + 100*time.Millisecond)

	assert.True(t, dl.TryLock(), "Expected to acquire lock after expiration")
}

func testMongoDistLock_ConcurrentLocking(t *testing.T) {
	lockID := "test-lock-concurrent"
	ttl := 2 * time.Second

	dl1 := NewMongoDistLock(db, lockID, ttl)
	dl2 := NewMongoDistLock(db, lockID, ttl)

	assert.True(t, dl1.TryLock(), "Expected to acquire lock by first instance")
	assert.False(t, dl2.TryLock(), "Expected second instance to fail to acquire lock")

	dl1.Unlock()
	assert.True(t, dl2.TryLock(), "Expected second instance to acquire lock after first instance unlocks")
}

func testMongoDistLock_LockRenew(t *testing.T) {
	lockID := "test-lock-renew"
	ttl := time.Second

	dl := NewMongoDistLock(db, lockID, ttl)

	assert.True(t, dl.TryLock(), "Expected to acquire lock")

	time.Sleep(ttl / 2)
	assert.False(t, dl.TryLock(), "Expected to fail to acquire lock while renew is active")

	time.Sleep(ttl)
	assert.True(t, dl.TryLock(), "Expected to acquire lock after ttl expires")
}

func testMongoDistLock_Quit(t *testing.T) {
	lockID := "test-lock-quit"
	ttl := time.Second
	dl1 := NewMongoDistLock(db, lockID, ttl)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	dl1.Lock(ctx)
	defer dl1.Unlock()

	db.Disconnect(ctx)
	time.Sleep(time.Second)

	for {
		select {
		case <-ctx.Done():
			assert.Fail(t, "Context timed out before ownership changed")
		case <-dl1.Quit():
			return
		}
	}
}
