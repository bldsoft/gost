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
	id  string
	gen uint64

	interval time.Duration
	do       func(ctx context.Context)
}

type IntervalRunner struct {
	queue        *timedqueue.TimedQueue[scheduledTask]
	wp           *WorkerPool
	gen          uint64
	taskToGen    map[string]uint64
	taskToCancel map[string]context.CancelFunc
	mtx          sync.Mutex
}

func NewIntervalRunner(workerN int) *IntervalRunner {
	return &IntervalRunner{
		queue:        timedqueue.New[scheduledTask](),
		wp:           new(WorkerPool).SetWorkerN(int64(workerN)),
		taskToGen:    make(map[string]uint64),
		taskToCancel: make(map[string]context.CancelFunc),
	}
}

func (r *IntervalRunner) SetWorkerN(workerN int) {
	r.wp.SetWorkerN(int64(workerN))
}

func (r *IntervalRunner) Add(id string, interval time.Duration, f func(ctx context.Context)) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	_, alreadyExists := r.taskToGen[id]
	r.cancelTaskContext(id)

	gen := r.nextGen()
	r.taskToGen[id] = gen
	task := scheduledTask{
		id:       id,
		gen:      gen,
		interval: interval,
		do:       f,
	}

	if alreadyExists {
		r.queue.RemoveFirstFunc(func(t scheduledTask) bool { return t.id == id })
	}
	r.queue.Push(task, time.Now())
}

func (r *IntervalRunner) Remove(id string) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	_, exists := r.taskToGen[id]
	if !exists {
		return
	}

	r.cancelTaskContext(id)
	r.queue.RemoveFirstFunc(func(t scheduledTask) bool { return t.id == id })
	delete(r.taskToGen, id)
}

func (r *IntervalRunner) Run(ctx context.Context) {
	for task := range r.queue.SyncSeq(ctx) {
		r.wp.In() <- func() {
			r.mtx.Lock()
			taskCtx, taskCancel := context.WithCancel(ctx)
			r.taskToCancel[task.id] = taskCancel
			r.mtx.Unlock()

			defer func() {
				taskCancel()
				r.mtx.Lock()
				defer r.mtx.Unlock()
				gen, exists := r.taskToGen[task.id]
				if !exists {
					return
				}

				if gen != task.gen {
					return
				} else {
					delete(r.taskToCancel, task.id)
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

			task.do(taskCtx)
		}
	}
	r.wp.CloseAndWait()
}

func (r *IntervalRunner) nextGen() uint64 {
	r.gen++
	return r.gen
}

func (r *IntervalRunner) cancelTaskContext(id string) {
	if cancel, ok := r.taskToCancel[id]; ok {
		cancel()
		delete(r.taskToCancel, id)
	}
}
