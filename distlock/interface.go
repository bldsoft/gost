package distlock

import "context"

type DistrMutex interface {
	Lock(ctx context.Context)
	TryLock() bool
	Unlock()
	Quit() <-chan struct{}
}
