package test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	workerpool "github.com/bldsoft/gost/utils/worker_pool"
	"github.com/stretchr/testify/require"
)

func TestWorkerPool_ChangeWorkerN(t *testing.T) {
	var wp workerpool.WorkerPool

	var workerN atomic.Int64
	addTasks := func(n int) (release func()) {
		releaseC := make(chan struct{})
		for range n {
			wp.In() <- func() {
				workerN.Add(1)
				defer workerN.Add(-1)
				<-releaseC
			}
		}
		return func() {
			close(releaseC)
		}
	}

	t.Run("increase workers", func(t *testing.T) {
		release := addTasks(30)

		wp.SetWorkerN(10)
		time.Sleep(time.Millisecond)
		require.Equal(t, int64(10), workerN.Load())
		require.Equal(t, int64(10), wp.WorkerN())

		wp.SetWorkerN(20)
		time.Sleep(time.Millisecond)
		require.Equal(t, int64(20), workerN.Load())

		release()
	})

	t.Run("decrease workers", func(t *testing.T) {
		wp.SetWorkerN(20)
		release := addTasks(20)

		time.Sleep(time.Millisecond)
		require.Equal(t, int64(20), workerN.Load())

		wp.SetWorkerN(10)
		release()

		time.Sleep(time.Millisecond)
		release = addTasks(20)
		time.Sleep(time.Millisecond * 10)
		require.Equal(t, int64(10), workerN.Load())

		release()
	})

	wp.CloseAndWait()
}

func TestWorkerPool_AllTasksDoneOnClose(t *testing.T) {
	var wg workerpool.WorkerPool
	var n atomic.Int64
	const taskN = 1000

	wg.SetWorkerN(1)
	for range taskN {
		wg.In() <- func() {
			time.Sleep(1 * time.Millisecond)
			n.Add(1)
		}
	}
	wg.CloseAndWait()
	require.Equal(t, int64(taskN), n.Load())
}

func TestWorkerPool_Group(t *testing.T) {
	var wp workerpool.WorkerPool
	wp.SetWorkerN(10)

	t.Run("error", func(t *testing.T) {
		g := wp.Group(context.Background())
		expectedErr := errors.New("err")
		g.Submit(func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
		g.Submit(func(ctx context.Context) error {
			return expectedErr
		})
		err := g.Wait()
		require.Equal(t, expectedErr, err)
	})

	t.Run("panic", func(t *testing.T) {
		g := wp.Group(context.Background())
		g.Submit(func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
		g.Submit(func(ctx context.Context) error {
			panic("panic")
		})
		err := g.Wait()
		require.Error(t, err)
		require.NotErrorIs(t, err, context.Canceled)
	})

	t.Run("context canceled", func(t *testing.T) {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		g := wp.Group(ctx)
		g.Submit(func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
		cancel()
		err := g.Wait()
		require.ErrorIs(t, err, context.Canceled)
	})
}
