package test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"

	workerpool "github.com/bldsoft/gost/utils/worker_pool"
	"github.com/stretchr/testify/require"
)

type wpTestFixture struct {
	wp      workerpool.WorkerPool
	workerN atomic.Int64
}

func (t *wpTestFixture) addTasks(n int) (release func()) {
	releaseC := make(chan struct{})
	for range n {
		t.wp.In() <- func() {
			t.workerN.Add(1)
			defer t.workerN.Add(-1)
			<-releaseC
		}
	}
	return func() {
		close(releaseC)
	}
}

func TestWorkerPool_ChangeWorkerN(t *testing.T) {

	t.Run("increase workers", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var f wpTestFixture
			defer f.wp.CloseAndWait()

			release := f.addTasks(30)
			defer release()

			f.wp.SetWorkerN(10)
			synctest.Wait()
			require.Equal(t, int64(10), f.workerN.Load())
			require.Equal(t, int64(10), f.wp.WorkerN())

			f.wp.SetWorkerN(20)
			synctest.Wait()
			require.Equal(t, int64(20), f.workerN.Load())

		})
	})

	t.Run("decrease workers", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var f wpTestFixture
			defer f.wp.CloseAndWait()

			f.wp.SetWorkerN(20)
			func() {
				release := f.addTasks(20)
				defer release()

				synctest.Wait()
				require.Equal(t, int64(20), f.workerN.Load())

				f.wp.SetWorkerN(10)
			}()

			synctest.Wait()
			release := f.addTasks(20)
			defer release()
			synctest.Wait()
			require.Equal(t, int64(10), f.workerN.Load())
		})
	})
}

func TestWorkerPool_AllTasksDoneOnClose(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var wp workerpool.WorkerPool
		var n atomic.Int64
		const taskN = 1000

		wp.SetWorkerN(1)
		for range taskN {
			wp.In() <- func() {
				synctest.Wait()
				n.Add(1)
			}
		}
		wp.CloseAndWait()
		require.Equal(t, int64(taskN), n.Load())
	})
}

func TestWorkerPool_Group(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var wp workerpool.WorkerPool
			defer wp.CloseAndWait()
			wp.SetWorkerN(10)

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
	})

	t.Run("panic", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var wp workerpool.WorkerPool
			defer wp.CloseAndWait()
			wp.SetWorkerN(10)

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
	})

	t.Run("context canceled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var wp workerpool.WorkerPool
			defer wp.CloseAndWait()
			wp.SetWorkerN(10)

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
	})
}
