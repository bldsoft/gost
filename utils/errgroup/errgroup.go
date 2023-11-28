package errgroup

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
)

type Group struct {
	wg  sync.WaitGroup
	err uniqErr
	mut sync.Mutex
	pnc atomic.Pointer[panicWrapper]
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.recover()
		defer g.wg.Done()

		if _err := f(); _err != nil {
			g.mut.Lock()
			g.err.add(_err)
			g.mut.Unlock()
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()

	if pnc := g.pnc.Load(); pnc != nil {
		panic(pnc)
	}

	return g.err
}

func (g *Group) recover() {
	if val := recover(); val != nil {
		p := NewPanic(1, val)
		g.pnc.Store(&p)
	}
}

type panicWrapper struct {
	Value   any
	Callers []uintptr
	Stack   []byte
}

func (w *panicWrapper) Error() string {
	return fmt.Sprintf("panic: %v\nstacktrace:\n%s\n", w.Value, w.Stack)
}

func NewPanic(skip int, value any) panicWrapper {
	var callers [64]uintptr
	n := runtime.Callers(skip, callers[:])
	return panicWrapper{
		Value:   value,
		Callers: callers[:n],
		Stack:   debug.Stack(),
	}
}

type uniqErr map[string]struct{}

func (ue uniqErr) add(err error) {
	ue[err.Error()] = struct{}{}
}

func (ue uniqErr) Error() string {
	var sb strings.Builder
	for err := range ue {
		sb.WriteString(err)
		sb.WriteByte(' ')
	}
	return sb.String()
}
