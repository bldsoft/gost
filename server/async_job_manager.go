package server

import (
	"context"
)

type AsyncRunner interface {
	Run() error
	Stop(ctx context.Context) error
}

type AsyncJobManager struct {
	jobGroup AsyncJobGroup
}

func NewRunnerManager(runners ...AsyncRunner) *AsyncJobManager {
	return &AsyncJobManager{AsyncJobGroup{runners}}
}

func (m *AsyncJobManager) Append(runners ...AsyncRunner) {
	m.jobGroup.Append(runners...)
}

func (m *AsyncJobManager) Start() {
	m.jobGroup.Run()
}

func (m *AsyncJobManager) Stop(ctx context.Context) error {
	return m.jobGroup.Stop(ctx)
}
