package workerpool

import (
	"context"
	"sync"

	"github.com/bldsoft/gost/utils/errgroup"
)

type workerPool interface {
	In() chan<- func()
}

type Group struct {
	wp        *WorkerPool
	ctx       context.Context
	cancel    func()
	waitGroup sync.WaitGroup

	err     error
	errOnce sync.Once
}

func newGroup(wp *WorkerPool, ctx context.Context) *Group {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{
		wp:     wp,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (g *Group) Submit(f func(ctx context.Context) error) {
	g.waitGroup.Add(1)
	g.wp.In() <- func() {
		defer g.recover()
		defer g.waitGroup.Done()

		select {
		case <-g.ctx.Done():
			g.setError(g.ctx.Err())
			return
		default:
		}

		if err := f(g.ctx); err != nil {
			g.setError(err)
		}
	}
}

func (g *Group) recover() {
	if val := recover(); val != nil {
		p := errgroup.NewPanic(1, val)
		g.setError(&p)
	}
}

func (g *Group) setError(err error) {
	g.errOnce.Do(func() {
		g.err = err
		g.cancel()
	})
}

func (g *Group) Wait() error {
	g.waitGroup.Wait()
	return g.err
}
