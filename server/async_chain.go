package server

import "slices"

import "context"

// AsyncJobChain starts jobs in sequence and stops them in reverse order.
type AsyncJobChain struct {
	runners []AsyncRunner
}

func NewAsyncJobChain(runners ...AsyncRunner) *AsyncJobChain {
	return &AsyncJobChain{runners}
}

func (c *AsyncJobChain) Append(runner AsyncRunner) {
	c.runners = append(c.runners, runner)
}

func (c *AsyncJobChain) Run() error {
	for _, r := range c.runners {
		if err := r.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (c *AsyncJobChain) Stop(ctx context.Context) error {
	for _, v := range slices.Backward(c.runners) {
		if err := v.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}
