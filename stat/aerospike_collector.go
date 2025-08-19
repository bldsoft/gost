package stat

import (
	"context"

	"github.com/bldsoft/gost/cache/v2/aerospike"
)

type AerospikeCollector struct {
	cache *aerospike.Storage
}

func NewAerospikeCollector(cache *aerospike.Storage) *AerospikeCollector {
	return &AerospikeCollector{cache: cache}
}

func (c *AerospikeCollector) Stat(ctx context.Context) Stat {
	stat, err := c.cache.Stat()
	return NewStat("aerospike", stat, err)
}
