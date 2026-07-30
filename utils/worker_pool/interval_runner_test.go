package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIntervalRunner_BasicExecution(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int64
		runner := NewIntervalRunner(2)

		runner.Add("task1", 10*time.Second, func(ctx context.Context) {
			count.Add(1)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		time.Sleep(25 * time.Second)
		cancel()
		time.Sleep(time.Second)

		got := count.Load()
		require.GreaterOrEqual(t, got, int64(3), "expected at least 3 executions, got %d", got)
	})
}

func TestIntervalRunner_Remove(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int64
		runner := NewIntervalRunner(2)

		runner.Add("task1", 5*time.Second, func(ctx context.Context) {
			count.Add(1)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		time.Sleep(2 * time.Second)
		runner.Remove("task1")
		time.Sleep(20 * time.Second)

		cancel()
		time.Sleep(time.Second)

		got := count.Load()
		require.Equal(t, int64(1), got, "task should have executed once before removal")
	})
}

func TestIntervalRunner_AddReplace(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var countA, countB atomic.Int64
		runner := NewIntervalRunner(2)

		runner.Add("task1", 5*time.Second, func(ctx context.Context) {
			countA.Add(1)
		})

		runner.Add("task1", 5*time.Second, func(ctx context.Context) {
			countB.Add(1)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		time.Sleep(2 * time.Second)
		cancel()
		time.Sleep(time.Second)

		require.Equal(t, int64(0), countA.Load(), "old function should not have run")
		require.Equal(t, int64(1), countB.Load(), "new function should have run")
	})
}

func TestIntervalRunner_ContextCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int64
		runner := NewIntervalRunner(2)

		runner.Add("task1", 5*time.Second, func(ctx context.Context) {
			count.Add(1)
		})

		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		defer cancel()

		runner.Run(ctx)

		got := count.Load()
		require.GreaterOrEqual(t, got, int64(2))
		require.LessOrEqual(t, got, int64(3))
	})
}

func TestIntervalRunner_RemoveDuringExecution(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int64
		runner := NewIntervalRunner(2)

		runner.Add("task1", 5*time.Second, func(ctx context.Context) {
			count.Add(1)
			time.Sleep(3 * time.Second)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		time.Sleep(2 * time.Second)
		runner.Remove("task1")

		time.Sleep(20 * time.Second)
		cancel()
		time.Sleep(time.Second)

		got := count.Load()
		require.Equal(t, int64(1), got, "task should not be rescheduled after removal during execution")
	})
}

func TestIntervalRunner_PanicRecovery(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int64
		runner := NewIntervalRunner(2)

		runner.Add("panicker", 5*time.Second, func(ctx context.Context) {
			count.Add(1)
			panic("test panic")
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		time.Sleep(12 * time.Second)
		cancel()
		time.Sleep(time.Second)

		got := count.Load()
		require.GreaterOrEqual(t, got, int64(2), "panicking task should be rescheduled")
	})
}

func TestIntervalRunner_ConcurrentAddRemove(t *testing.T) {
	runner := NewIntervalRunner(4)
	ctx, cancel := context.WithCancel(context.Background())

	var executions sync.Map

	go runner.Run(ctx)

	var wg sync.WaitGroup
	const goroutines = 10
	const iterations = 100

	for g := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := "shared-task"
			for i := range iterations {
				if (g+i)%3 == 0 {
					runner.Remove(id)
				} else {
					runner.Add(id, time.Millisecond, func(ctx context.Context) {
						executions.Store("ran", true)
					})
				}
			}
		}()
	}

	wg.Wait()
	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)
}

func TestIntervalRunner_MultipleTasks(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var countA, countB atomic.Int64
		runner := NewIntervalRunner(4)

		runner.Add("a", 5*time.Second, func(ctx context.Context) {
			countA.Add(1)
		})
		runner.Add("b", 10*time.Second, func(ctx context.Context) {
			countB.Add(1)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		time.Sleep(21 * time.Second)
		cancel()
		time.Sleep(time.Second)

		gotA := countA.Load()
		gotB := countB.Load()
		require.GreaterOrEqual(t, gotA, int64(4), "task A with 5s interval should run ~4 times in 21s")
		require.GreaterOrEqual(t, gotB, int64(2), "task B with 10s interval should run ~2 times in 21s")
	})
}

func TestIntervalRunner_AddRemoveAddDuringExecution(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var countA, countB atomic.Int64
		runner := NewIntervalRunner(4)

		runner.Add("task1", 5*time.Second, func(ctx context.Context) {
			countA.Add(1)
			time.Sleep(5 * time.Second)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		time.Sleep(1 * time.Second)
		runner.Add("task1", 5*time.Second, func(ctx context.Context) { countA.Add(100) })
		runner.Remove("task1")
		runner.Add("task1", 5*time.Second, func(ctx context.Context) {
			countB.Add(1)
		})

		time.Sleep(20 * time.Second)
		cancel()
		time.Sleep(time.Second)

		require.Equal(t, int64(1), countA.Load(), "original task should have completed once")
		require.GreaterOrEqual(t, countB.Load(), int64(1), "last-added task should eventually run")
	})
}

func TestIntervalRunner_LastWins_RemoveThenAdd(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int64
		runner := NewIntervalRunner(2)

		runner.Add("task1", 5*time.Second, func(ctx context.Context) {
			count.Add(1)
		})
		runner.Remove("task1")
		runner.Add("task1", 5*time.Second, func(ctx context.Context) {
			count.Add(10)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		time.Sleep(2 * time.Second)
		cancel()
		time.Sleep(time.Second)

		got := count.Load()
		require.Equal(t, int64(10), got, "last Add should win after Remove")
	})
}

func TestIntervalRunner_ContextCancelledOnRemove(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		runner := NewIntervalRunner(1)
		started := make(chan struct{})
		unblocked := make(chan struct{})

		runner.Add("task1", time.Second, func(ctx context.Context) {
			close(started)
			<-ctx.Done()
			close(unblocked)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		<-started
		runner.Remove("task1")

		select {
		case <-unblocked:
		case <-time.After(2 * time.Second):
			t.Fatal("task context was not cancelled on Remove")
		}

		cancel()
		synctest.Wait()
	})
}

func TestIntervalRunner_ContextCancelledOnReplace(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		runner := NewIntervalRunner(1)
		oldStarted := make(chan struct{})
		oldUnblocked := make(chan struct{})
		newStarted := make(chan struct{})
		var newContextFresh atomic.Bool

		runner.Add("task1", time.Second, func(ctx context.Context) {
			close(oldStarted)
			<-ctx.Done()
			close(oldUnblocked)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		<-oldStarted
		runner.Add("task1", time.Second, func(ctx context.Context) {
			newContextFresh.Store(ctx.Err() == nil)
			close(newStarted)
		})

		select {
		case <-oldUnblocked:
		case <-time.After(2 * time.Second):
			t.Fatal("old task context was not cancelled on replace")
		}

		select {
		case <-newStarted:
		case <-time.After(2 * time.Second):
			t.Fatal("new task did not start after replace")
		}

		require.True(t, newContextFresh.Load(), "new task should receive a non-cancelled context")

		cancel()
		synctest.Wait()
	})
}

func TestIntervalRunner_ContextCancelledOnShutdown(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		runner := NewIntervalRunner(1)
		started := make(chan struct{})
		unblocked := make(chan struct{})

		runner.Add("task1", time.Second, func(ctx context.Context) {
			close(started)
			<-ctx.Done()
			close(unblocked)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		<-started
		cancel()

		select {
		case <-unblocked:
		case <-time.After(2 * time.Second):
			t.Fatal("task context was not cancelled on runner shutdown")
		}
	})
}

func TestIntervalRunner_NewTaskGetsOwnContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		runner := NewIntervalRunner(1)
		var newContextFresh atomic.Bool
		checked := make(chan struct{})

		runner.Add("task1", time.Second, func(ctx context.Context) {
			<-ctx.Done()
		})
		runner.Add("task1", time.Second, func(ctx context.Context) {
			newContextFresh.Store(ctx.Err() == nil)
			close(checked)
		})

		ctx, cancel := context.WithCancel(context.Background())
		go runner.Run(ctx)

		select {
		case <-checked:
		case <-time.After(2 * time.Second):
			t.Fatal("replaced task did not run")
		}

		require.True(t, newContextFresh.Load(), "replaced task should receive a fresh non-cancelled context")

		cancel()
		synctest.Wait()
	})
}

func TestIntervalRunner_AddAfterRun(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var count atomic.Int64
		runner := NewIntervalRunner(1)
		started := make(chan struct{})

		runner.Add("setup", time.Hour, func(ctx context.Context) {
			close(started)
		})

		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() {
			runner.Run(ctx)
			close(done)
		}()

		<-started
		cancel()
		<-done

		runner.Add("task1", time.Second, func(ctx context.Context) {
			count.Add(1)
		})

		time.Sleep(2 * time.Second)
		require.Equal(t, int64(0), count.Load(), "task added after Run exits should not execute")
	})
}
