package healthcheck

import (
	"context"

	"github.com/bldsoft/gost/cache/memcached"
)

type MemcachedHealthChecker struct {
	db *memcached.Storage
}

func NewMemcachedHealthChecker(db *memcached.Storage) *MemcachedHealthChecker {
	return &MemcachedHealthChecker{db: db}
}

func (c *MemcachedHealthChecker) CheckHealth(ctx context.Context) Health {
	stats, err := c.db.Stats(ctx)
	return NewHealth("memcached", stats, err)
}
