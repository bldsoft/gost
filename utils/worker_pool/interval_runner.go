package workerpool

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/timedqueue"
)

type scheduledTask struct {
	id string

	interval time.Duration
	do       func(ctx context.Context)

	ctx    context.Context
	cancel context.CancelFunc
}

func newScheduledTask(id string, interval time.Duration, f func(ctx context.Context)) *scheduledTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &scheduledTask{
		id:       id,
		interval: interval,
		do:       f,
		ctx:      ctx,
		cancel:   cancel,
	}
}

type IntervalRunner struct {
	queue    *timedqueue.TimedQueue[*scheduledTask]
	wp       *WorkerPool
	idToTask map[string]*scheduledTask
	mtx      sync.Mutex
}

func NewIntervalRunner(workerN int) *IntervalRunner {
	return &IntervalRunner{
		queue:    timedqueue.New[*scheduledTask](),
		wp:       new(WorkerPool).SetWorkerN(int64(workerN)),
		idToTask: make(map[string]*scheduledTask),
	}
}

func (r *IntervalRunner) SetWorkerN(workerN int) {
	r.wp.SetWorkerN(int64(workerN))
}

func (r *IntervalRunner) Add(id string, interval time.Duration, f func(ctx context.Context)) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if existingTask := r.idToTask[id]; existingTask != nil {
		r.queue.RemoveFirstFunc(func(t *scheduledTask) bool { return t.id == id })
		existingTask.cancel()
	}

	task := newScheduledTask(id, interval, f)
	r.idToTask[id] = task
	r.queue.Push(task, time.Now())
}

func (r *IntervalRunner) Remove(id string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	task, exists := r.idToTask[id]
	if !exists {
		return
	}

	r.queue.RemoveFirstFunc(func(t *scheduledTask) bool { return t.id == id })
	task.cancel()
	delete(r.idToTask, id)
}

func (r *IntervalRunner) Run(ctx context.Context) {
	for task := range r.queue.SyncSeq(ctx) {
		r.wp.In() <- func() {
			defer func() {
				r.mtx.Lock()
				defer r.mtx.Unlock()
				currentTask, exists := r.idToTask[task.id]
				if !exists {
					return
				}
				if currentTask != task {
					return
				}
				r.queue.Push(task, time.Now().Add(task.interval))
			}()
			defer func() {
				if err := recover(); err != nil {
					log.ErrorfWithFields(log.Fields{
						"error":       err,
						"stack trace": string(debug.Stack()),
					}, "IntervalRunner task panicked")
				}
			}()
			task.do(task.ctx)
		}
	}

	r.mtx.Lock()
	for _, task := range r.idToTask {
		task.cancel()
	}
	r.mtx.Unlock()
	r.wp.CloseAndWait()
}
