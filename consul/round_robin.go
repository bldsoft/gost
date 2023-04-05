package consul

import (
	"sync/atomic"
)

type RoundRobin struct {
	offset atomic.Int64
}

func (rr *RoundRobin) Balance(serviceCluster string, addrs []string) []string {
	if len(addrs) == 0 {
		return addrs
	}
	i := (rr.offset.Add(1) - 1) % int64(len(addrs))
	return append(addrs[i:], addrs[i+1:]...)
}
