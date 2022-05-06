package healthcheck

import (
	"context"

	"github.com/bldsoft/gost/clickhouse"
)

type ClickHouseChecker struct {
	db *clickhouse.Clickhouse
}

func NewClickHouseHealthChecker(db *clickhouse.Clickhouse) *ClickHouseChecker {
	return &ClickHouseChecker{db: db}
}

func (c *ClickHouseChecker) CheckHealth(ctx context.Context) Health {
	stat, err := c.db.Stats(ctx)
	return NewHealth("clickhouse", stat, err)
}
