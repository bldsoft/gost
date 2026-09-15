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

	"github.com/stretchr/testify/assert"

	"github.com/bldsoft/gost/cache/v2"
	"github.com/bldsoft/gost/log"
)

const lockKey = "lock:test:cache-v2"

var (
	rep *MemcacheRepository
)

func TestMain(m *testing.M) {
	log.SetLogLevel("")
	storage := NewStorage([]string{"127.0.0.1:11211"}, Config{})
	rep = NewMemcacheRepository(storage, 2*time.Minute)

	exitVal := m.Run()
	os.Exit(exitVal)
}

func incrCounter(t *testing.T, id string, counter *int32) {
	n := atomic.AddInt32(counter, 1)
	t.Logf("%s enter lock. counter = %d", id, n)
}

func decrCounter(t *testing.T, id string, counter *int32) {
	n := atomic.AddInt32(counter, -1)
	t.Logf("%s exit lock. counter = %d", id, n)
}

func routine(ctx context.Context, t *testing.T, key, id string, counter *int32, unlockTime time.Duration, stopGoroutine <-chan struct{}) {
	mtx := cache.NewDistrMutex(rep, key, unlockTime)
	mtx.TryLockInterval = 50 * time.Millisecond
	for {
		if ctx.Err() != nil {
			return
		}
		t.Logf("%s is waiting for lock", id)
		mtx.Lock(ctx)
		if ctx.Err() != nil {
			mtx.Unlock()

			return
		}
		incrCounter(t, id, counter)
	SelectLoop:
		for {
			select {
			case <-time.After(unlockTime):
				t.Logf("%s tick", id)
			case <-mtx.Quit():
				t.Logf("%s replaced", id)
				break SelectLoop
			case <-stopGoroutine:
				t.Logf("%s is killed", id)
				decrCounter(t, id, counter)
				mtx.Unlock()

				return
			case <-ctx.Done():
				decrCounter(t, id, counter)
				mtx.Unlock()

				return
			}
		}
		decrCounter(t, id, counter)
		mtx.Unlock()
	}
}

func testTemplate(t *testing.T, f func(key string, unlockTime time.Duration, stopGoroutine chan struct{})) {
	key := lockKey + ":" + t.Name()
	_ = rep.cache.Delete(key)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	const n = 3
	stopGoroutine := make(chan struct{})
	unlockTime := time.Second
	var counter int32
	for i := 1; i <= n; i++ {
		go routine(ctx, t, key, fmt.Sprintf("%d", i), &counter, unlockTime, stopGoroutine)
	}

	time.Sleep(200 * time.Millisecond)
	f(key, unlockTime, stopGoroutine)
	assert.Eventually(t, func() bool {
		return atomic.LoadInt32(&counter) == 1
	}, 5*time.Second, 50*time.Millisecond)

	t.Log("-- ending test --")
	cancel()
	time.Sleep(200 * time.Millisecond)
}

func TestLock(t *testing.T) {
	testTemplate(t, func(_ string, unlockTime time.Duration, _ chan struct{}) {
		time.Sleep(unlockTime * 3)
	})
}

func TestLockOwnerKilledCase(t *testing.T) {
	testTemplate(t, func(_ string, unlockTime time.Duration, stopGoroutine chan struct{}) {
		stopGoroutine <- struct{}{}
		time.Sleep(unlockTime * 3)
	})
}

func TestLockDeleting(t *testing.T) {
	testTemplate(t, func(key string, unlockTime time.Duration, _ chan struct{}) {
		for range 10 {
			_ = rep.cache.Delete(key)
			time.Sleep(100 * time.Millisecond)
		}
		time.Sleep(unlockTime * 3)
	})
}

func TestAtomicIncrement(t *testing.T) {
	const goroutineN = 10
	const lockCount = 200

	key := lockKey + ":" + t.Name()
	_ = rep.cache.Delete(key)

	var wg, barier sync.WaitGroup
	wg.Add(goroutineN)
	barier.Add(1)

	var i int32
	increment := func(n int) {
		defer wg.Done()

		mtx := cache.NewDistrMutex(rep, key, time.Minute)
		mtx.TryLockInterval = time.Millisecond

		wg.Done()
		barier.Wait()

		for range n {
			mtx.Lock(context.Background())
			atomic.AddInt32(&i, 1)
			mtx.Unlock()
		}
	}

	for range goroutineN {
		go increment(lockCount)
	}

	wg.Wait()
	wg.Add(goroutineN)
	barier.Done()
	wg.Wait()

	assert.Equal(t, int32(goroutineN*lockCount), atomic.LoadInt32(&i))
}
