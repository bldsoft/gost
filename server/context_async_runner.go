package server

import (
	"context"
)

func NewContextAsyncRunner(run func(context.Context) error) AsyncRunner {
	return &contextAsyncRunner{
		run: run,
	}
}

type contextAsyncRunner struct {
	run     func(context.Context) error
	ctx     context.Context
	stop    func()
	stopped chan struct{}
}

func (r *contextAsyncRunner) Run() error {
	r.ctx, r.stop = context.WithCancel(context.Background())
	r.stopped = make(chan struct{})
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
