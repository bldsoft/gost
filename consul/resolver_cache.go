package consul

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrCacheMiss = errors.New("cache miss")

type cacheEntry struct {
	addrs []string
	exp   time.Time
}

type resolverCache struct {
	cache sync.Map // cluster name -> { addrs, exp time }
	ttl   time.Duration
}

func newResolverCache(ttl time.Duration) *resolverCache {
	return &resolverCache{ttl: ttl}
}

func (c *resolverCache) cacheTTL() time.Duration {
	if c.ttl == 0 {
		return 5 * time.Minute
	}
	return c.ttl
}

func (c *resolverCache) put(serviceCluster string, addrs []string) {
	c.cache.Store(serviceCluster, cacheEntry{addrs, time.Now().Add(c.cacheTTL())})
}

func (c *resolverCache) lookupServices(ctx context.Context, serviceCluster string) ([]string, error) {
	valI, ok := c.cache.Load(serviceCluster)
	if !ok {
		return nil, ErrCacheMiss
	}
	val := valI.(cacheEntry)

	if val.exp.Before(time.Now()) {
		return nil, ErrCacheMiss
	}

	return val.addrs, nil
}
