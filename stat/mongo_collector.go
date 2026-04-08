package stat

import (
	"context"

	"github.com/bldsoft/gost/mongo"
)

type MongoStatsProvider interface {
	Stats(ctx context.Context) (interface{}, error)
}

type MongoCollector struct {
	db MongoStatsProvider
}

func NewMongoCollector(db *mongo.Storage) *MongoCollector {
	return &MongoCollector{db: db}
}

func (c *MongoCollector) Stat(ctx context.Context) Stat {
	stat, err := c.db.Stats(ctx)
	return NewStat("mongo", stat, err)
}
