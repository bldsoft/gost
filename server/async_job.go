package server

import "context"

type AsyncJob struct {
	run  func() error
	stop func(ctx context.Context) error
}

func NewAsyncJob(run func() error, stop func(ctx context.Context) error) *AsyncJob {
	return &AsyncJob{run, stop}
}

func (j *AsyncJob) Run() error {
	if j.run == nil {
		return nil
	}
	return j.run()
}

func (j *AsyncJob) Stop(ctx context.Context) error {
	if j.stop == nil {
		return nil
	}
	return j.stop(ctx)
}
