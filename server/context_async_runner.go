package server

import (
	"context"
)

func NewContextAsyncRunner(run func(context.Context) error) AsyncRunner {
	return &contextAsyncRunner{
		run: run,
	}
}

type CauseStoppableRunner interface {
	AsyncRunner
	StopWithCause(ctx context.Context, cause error) error
}

type contextAsyncRunner struct {
	run func(context.Context) error

	ctx     context.Context
	stop    func(error)
	stopped chan struct{}
}

func (r *contextAsyncRunner) Run() error {
	r.ctx, r.stop = context.WithCancelCause(context.Background())
	r.stopped = make(chan struct{})
	defer close(r.stopped)
	return r.run(r.ctx)
}

func (r *contextAsyncRunner) Stop(ctx context.Context) error {
	return r.StopWithCause(ctx, context.Canceled)
}

func (r *contextAsyncRunner) StopWithCause(ctx context.Context, cause error) error {
	r.stop(cause)

	select {
	case <-r.stopped:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
