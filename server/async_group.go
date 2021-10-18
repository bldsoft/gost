package server

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/bldsoft/gost/log"
	"github.com/hashicorp/go-multierror"
)

// AsyncJobGroup runs jobs in parallel
type AsyncJobGroup struct {
	runners []AsyncRunner
}

func NewAsyncJobGroup(runners ...AsyncRunner) *AsyncJobGroup {
	return &AsyncJobGroup{runners}
}

func (c *AsyncJobGroup) Append(runner ...AsyncRunner) {
	c.runners = append(c.runners, runner...)
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else {
		return t.Name()
	}
}
func errFormat(errors []error) string {
	result := fmt.Sprintf("%d error occurred", len(errors))
	for _, err := range errors {
		result += "\n"
		result += fmt.Sprintf("* %v", err)
	}
	return result
}
func (m *AsyncJobGroup) runParallel(f func(r AsyncRunner) error) error {
	errC := make(chan error, len(m.runners))
	go func() {
		defer close(errC)
		var wg sync.WaitGroup
		wg.Add(len(m.runners))
		for _, runner := range m.runners {
			go func(r AsyncRunner) {
				if err := f(r); err != nil {
					errC <- fmt.Errorf("%s: %w", getType(r), err)
				}
				wg.Done()
			}(runner)
		}
		wg.Wait()
	}()

	var multiErr *multierror.Error
	for err := range errC {
		multiErr = multierror.Append(multiErr, err)
	}
	if multiErr == nil {
		return nil
	}
	multiErr.ErrorFormat = errFormat
	return multiErr
}

func (m *AsyncJobGroup) Run() error {
	return m.runParallel(func(r AsyncRunner) error {
		err := r.Run()
		log.DebugOrErrorf(err, "%s Run ended ", getType(r))
		return err
	})
}

func (m *AsyncJobGroup) Stop(ctx context.Context) error {
	return m.runParallel(func(r AsyncRunner) error {
		return r.Stop(ctx)
	})
}
