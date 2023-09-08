package discovery

import (
	"sync/atomic"
)

type RoundRobin[T any] struct {
	offset atomic.Int64
}

func (rr *RoundRobin[T]) Balance(serviceCluster string, addrs []T) []T {
	if len(addrs) == 0 {
		return addrs
	}
	i := (rr.offset.Add(1) - 1) % int64(len(addrs))
	return append(addrs[i:], addrs[i+1:]...)
}
