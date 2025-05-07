package storage

import (
	"sync/atomic"
)

type AsyncDB interface {
	NotifyReady() <-chan struct{}
}

type taskScheduler struct {
	active atomic.Bool
	taskCh chan func() error
}

var scheduler = &taskScheduler{
	taskCh: make(chan func() error),
}

func ScheduleTask(db AsyncDB, t func() error) {
	scheduler.scheduleTask(db, t)
	go func() {
		scheduler.run()
	}()
}

func (s *taskScheduler) scheduleTask(db AsyncDB, t func() error) {
	s.taskCh <- func() error {
		db.NotifyReady()
		return t()
	}
}

func (s *taskScheduler) run() {
	if !s.active.CompareAndSwap(false, true) {
		return
	}

	for task := range s.taskCh {
		task()
	}
}
