package storage

import (
	"sync"
	"sync/atomic"

	"github.com/bldsoft/gost/log"
)

type TaskScheduler struct {
	active atomic.Bool
	taskCh chan func() error
	wg     *sync.WaitGroup
}

func NewTaskScheduler(wg *sync.WaitGroup) *TaskScheduler {
	return &TaskScheduler{
		taskCh: make(chan func() error, 100),
		wg:     wg,
	}
}

func (s *TaskScheduler) ScheduleTask(task func() error) {
	s.wg.Add(1)
	s.taskCh <- task
	go func() {
		s.run()
	}()
}

func (s *TaskScheduler) run() {
	if !s.active.CompareAndSwap(false, true) {
		return
	}
	defer s.active.Store(false)

	for task := range s.taskCh {
		if err := task(); err != nil {
			log.Fatalf("scheduled task failed: %v", err)
		}
		s.wg.Done()
	}
}
