package storage

import (
	"log"
	"sync"
	"sync/atomic"
)

type AsyncDB interface {
	NotifyReady() <-chan struct{}
}

type taskScheduler struct {
	active atomic.Bool
	taskCh chan func() error
	wg     *sync.WaitGroup
}

var scheduler = &taskScheduler{
	taskCh: make(chan func() error, 1),
}

func ScheduleTask(db AsyncDB, t func() error) {
	scheduler.scheduleTask(db, t)
	go func() {
		scheduler.run()
	}()
}

func (s *taskScheduler) scheduleTask(db AsyncDB, t func() error) {
	s.taskCh <- func() error {
		<-db.NotifyReady()
		return t()
	}
}

func (s *taskScheduler) run() {
	if !s.active.CompareAndSwap(false, true) {
		return
	}
	defer s.active.Store(false)

	for task := range s.taskCh {
		if err := task(); err != nil {
			log.Fatalf("scheduled task failed: %v", err)
		}
	}
}
