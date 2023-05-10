package server

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/errgroup"
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

func (m *AsyncJobGroup) runParallel(f func(r AsyncRunner) error) error {
	var errGroup errgroup.Group
	for _, runner := range m.runners {
		r := runner
		errGroup.Go(func() error {
			return f(r)
		})
	}
	return errGroup.Wait()
}

func (m *AsyncJobGroup) Run() error {
	return m.runParallel(func(r AsyncRunner) error {
		defer func() {
			if e := recover(); e != nil {
				log.Errorf("%s job ended: %v\nstacktrace:\n%s", getType(r), e, strings.TrimSpace(string(debug.Stack())))
			}
		}()
		err := r.Run()
		log.DebugOrErrorf(err, "%s job ended ", getType(r))
		return err
	})
}
func (m *AsyncJobGroup) Stop(ctx context.Context) error {
	return m.runParallel(func(r AsyncRunner) error {
		if err := r.Stop(ctx); err != nil {
			return fmt.Errorf("%s: %w", getType(r), err)
		}
		return nil
	})
}
