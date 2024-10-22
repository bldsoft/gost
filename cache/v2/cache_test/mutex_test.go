//go:build integration_test

package memcached

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/stretchr/testify/assert"
)

const lockKey = "lock:test"

var (
	rep *MemcacheRepository
)

func TestMain(m *testing.M) {
	log.SetLogLevel("")
	storage := NewStorage(Config{[]string{"127.0.0.1:11211"}})
	rep = NewMemcacheRepository(storage, 2)

	exitVal := m.Run()
	os.Exit(exitVal)
}

func incrCounter(t *testing.T, id string, counter *int32) {
	atomic.AddInt32(counter, 1)
	t.Logf("%s enter lock. counter = %d", id, *counter)
}

func decrCounter(t *testing.T, id string, counter *int32) {
	atomic.AddInt32(counter, -1)
	t.Logf("%s exit lock. counter = %d", id, *counter)
}

func routine(t *testing.T, id string, counter *int32, unlockTime time.Duration, stopGoroutine chan struct{}) {
	mtx := NewMemcachedMutex(rep, lockKey, unlockTime)
	for {
		t.Logf("%s is waiting for lock", id)
		mtx.Lock(context.Background())
		defer mtx.Unlock()
		incrCounter(t, id, counter)
	SelectLoop:
		for {
			select {
			case <-time.Tick(unlockTime):
				t.Logf("%s tick", id)
			case <-mtx.Quit():
				t.Logf("%s replaced", id)
				break SelectLoop
			case <-stopGoroutine:
				t.Logf("%s is killed", id)
				decrCounter(t, id, counter)
				return
			}
		}
		decrCounter(t, id, counter)
	}
}

func testTemplate(t *testing.T, f func(unlockTime time.Duration, goroutineN int, stopGoroutine chan struct{})) {
	rep.cache.Delete(lockKey)
	const n = 3
	stopGoroutine := make(chan struct{})
	unlockTime := 10 * time.Millisecond
	var counter int32
	for i := 1; i <= n; i++ {
		go routine(t, fmt.Sprintf("%d", i), &counter, unlockTime, stopGoroutine)
	}

	time.Sleep(10 * time.Millisecond)
	f(unlockTime, n, stopGoroutine)
	assert.Equal(t, int32(1), counter)

	t.Log("-- ending test --")
	close(stopGoroutine)
	time.Sleep(unlockTime * n)
}

func TestLock(t *testing.T) {
	testTemplate(t, func(unlockTime time.Duration, _ int, _ chan struct{}) {
		time.Sleep(unlockTime * 3)
	})
}

func TestLockOwnerKilledCase(t *testing.T) {
	testTemplate(t, func(unlockTime time.Duration, _ int, stopGoroutine chan struct{}) {
		stopGoroutine <- struct{}{}
		time.Sleep(unlockTime * 3)
	})
}

func TestLockDeleting(t *testing.T) {
	testTemplate(t, func(unlockTime time.Duration, _ int, _ chan struct{}) {
		for i := 0; i < 10; i++ {
			rep.cache.Delete(lockKey)
			time.Sleep(unlockTime)
		}
		time.Sleep(unlockTime * 3)
	})
}

func TestAtomicIncrement(t *testing.T) {
	const goroutineN = 10
	const lockCount = 1000

	var wg, barier sync.WaitGroup
	wg.Add(goroutineN)
	barier.Add(1)

	var i int32
	var increment = func(n int) {
		defer wg.Done()

		mtx := NewMemcachedMutex(rep, lockKey, time.Second)
		mtx.TryLockInterval = time.Millisecond

		wg.Done()
		barier.Wait()

		for j := 0; j < n; j++ {
			mtx.Lock(context.TODO())
			i++
			mtx.Unlock()
			time.Sleep(time.Millisecond)
		}
	}

	for i := 0; i < goroutineN; i++ {
		go increment(lockCount)
	}

	wg.Wait()
	wg.Add(goroutineN)
	barier.Done()
	wg.Wait()

	assert.Equal(t, int32(goroutineN*lockCount), i)
}
