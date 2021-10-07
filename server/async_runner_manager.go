package server

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/bldsoft/gost/log"
)

type AsyncRunner interface {
	Run() error
	Stop(ctx context.Context) error
}

type AsyncRunnerManager struct {
	runners []AsyncRunner
}

func NewRunnerManager(runners ...AsyncRunner) *AsyncRunnerManager {
	return &AsyncRunnerManager{runners}
}

func (m *AsyncRunnerManager) Append(runners ...AsyncRunner) {
	m.runners = append(m.runners, runners...)
}

func (m *AsyncRunnerManager) Start() {
	for _, runner := range m.runners {
		go func(r AsyncRunner) {
			log.DebugOrErrorf(r.Run(), "%s Run ended ", getType(r))
		}(runner)
	}
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}

func (m *AsyncRunnerManager) Stop(ctx context.Context) chan error {
	errC := make(chan error)
	go func() {
		defer close(errC)
		var wg sync.WaitGroup
		wg.Add(len(m.runners))
		for _, runner := range m.runners {
			go func(r AsyncRunner) {
				if err := r.Stop(ctx); err != nil {
					errC <- fmt.Errorf("%s: %w", getType(r), err)
				}
				wg.Done()
			}(runner)
		}
		wg.Wait()
	}()

	return errC
}
