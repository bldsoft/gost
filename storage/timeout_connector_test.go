package storage

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDBConnectAsyncRetriesErrorAndPanic(t *testing.T) {
	var calls atomic.Int32
	var wg sync.WaitGroup
	DBConnectAsync(&wg, func() error {
		switch calls.Add(1) {
		case 1:
			return errors.New("down")
		case 2:
			panic("down")
		default:
			return nil
		}
	}, 3, time.Millisecond)
	wg.Wait()

	if calls.Load() != 3 {
		t.Fatalf("connect calls = %d, want 3", calls.Load())
	}
}
