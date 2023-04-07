package consul

import (
	"context"
	"errors"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/cache/bigcache"
)

var ErrCacheMiss = errors.New("cache miss")

type resolverCache struct {
	cache cache.Repository[[]string] // cluster name -> addrs
	ttl   time.Duration
}

func newResolverCache(ttl time.Duration) resolverCache {
	return resolverCache{ttl: ttl, cache: cache.Typed[[]string](bigcache.NewExpiringRepository("{}"))}
}

func (c *resolverCache) put(serviceCluster string, addrs []string) {
	c.cache.SetFor(serviceCluster, addrs, c.ttl)
}

func (c *resolverCache) lookupServices(ctx context.Context, serviceCluster string) ([]string, error) {
	return c.cache.Get(serviceCluster)
}
