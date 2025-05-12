package storage

import "sync/atomic"

type ReadyState struct {
	ready   atomic.Bool
	readyCh chan struct{}
}

func NewReadyState() *ReadyState {
	return &ReadyState{readyCh: make(chan struct{})}
}

func (r *ReadyState) SetReady() {
	r.ready.Store(true)
	close(r.readyCh)
}

func (r *ReadyState) IsReady() bool {
	return r.ready.Load()
}

func (r *ReadyState) NotifyReady() <-chan struct{} {
	return r.readyCh
}
