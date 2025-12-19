package workerpool

import (
	"context"
	"sync"
	"sync/atomic"
)

const (
	workPoolTaskChannelSize = 1 << 16
	workPoolMaxWorkers      = 1 << 16
)

type WorkerPool struct {
	workerN   int64
	taskC     chan func()
	taskCOnce sync.Once

	stopWorkerC chan struct{}

	wg sync.WaitGroup

	workerNMtx sync.RWMutex
}

func (wp *WorkerPool) WithTaskChannel(taskC chan func()) {
	wp.taskC = taskC
}

func (wp *WorkerPool) SetWorkerN(n int64) *WorkerPool {
	n = max(0, n)

	wp.workerNMtx.Lock()
	defer wp.workerNMtx.Unlock()

	if wp.stopWorkerC == nil {
		wp.stopWorkerC = make(chan struct{}, workPoolMaxWorkers)
	}

	if diff := n - wp.workerN; diff > 0 {
		for range diff {
			wp.startWorker()
		}
	} else {
		for range -diff {
			wp.stopWorkerC <- struct{}{}
		}
	}
	atomic.StoreInt64(&wp.workerN, n)
	return wp
}

func (wp *WorkerPool) WorkerN() int64 {
	return atomic.LoadInt64(&wp.workerN)
}

func (wp *WorkerPool) taskChan() chan func() {
	wp.taskCOnce.Do(func() {
		if wp.taskC == nil {
			wp.taskC = make(chan func(), workPoolTaskChannelSize)
		}
	})
	return wp.taskC
}

func (wp *WorkerPool) Group(ctx context.Context) *Group {
	return newGroup(wp, ctx)
}

// do not push on closed worker pool
func (wp *WorkerPool) In() chan<- func() {
	return wp.taskChan()
}

func (wp *WorkerPool) startWorker() {
	wp.wg.Add(1)

	go func() {
		defer wp.wg.Done()
		for {
			select {
			case <-wp.stopWorkerC:
				return
			case f, ok := <-wp.taskChan():
				if !ok {
					return
				}
				f()
			}
		}
	}()
}

func (wp *WorkerPool) CloseAndWait() {
	wp.Close()
	wp.Wait()
}

func (wp *WorkerPool) Close() {
	close(wp.taskC)
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()
}
