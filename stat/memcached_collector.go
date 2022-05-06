package stat

import (
	"context"

	"github.com/bldsoft/gost/cache/memcached"
)

type MemcachedCollector struct {
	db *memcached.Storage
}

func NewMemcachedCollector(db *memcached.Storage) *MemcachedCollector {
	return &MemcachedCollector{db: db}
}

func (c *MemcachedCollector) Stat(ctx context.Context) Stat {
	stats, err := c.db.Stats(ctx)
	return NewStat("memcached", stats, err)
}
