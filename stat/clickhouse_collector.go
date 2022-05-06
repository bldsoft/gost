package stat

import (
	"context"

	"github.com/bldsoft/gost/clickhouse"
)

type ClickHouseCollector struct {
	db *clickhouse.Storage
}

func NewClickHouseCollector(db *clickhouse.Storage) *ClickHouseCollector {
	return &ClickHouseCollector{db: db}
}

func (c *ClickHouseCollector) Stat(ctx context.Context) Stat {
	stat, err := c.db.Stats(ctx)
	return NewStat("clickhouse", stat, err)
}
