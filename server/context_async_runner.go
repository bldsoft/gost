package server

import (
	"context"
)

func NewContextAsyncRunner(run func(context.Context) error) AsyncRunner {
	ctx, cancel := context.WithCancel(context.Background())
	return &contextAsyncRunner{
		run:     run,
		ctx:     ctx,
		stop:    cancel,
		stopped: make(chan struct{}),
	}
}

type contextAsyncRunner struct {
	run     func(context.Context) error
	ctx     context.Context
	stop    func()
	stopped chan struct{}
}

func (r *contextAsyncRunner) Run() error {
	defer close(r.stopped)
	return r.run(r.ctx)
}

func (r *contextAsyncRunner) Stop(ctx context.Context) error {
	r.stop()
	select {
	case <-r.stopped:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
