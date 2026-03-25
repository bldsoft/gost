package stat

import (
	"context"

	"github.com/bldsoft/gost/mongo"
	mongov2 "github.com/bldsoft/gost/mongo/v2"
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

func NewMongoCollectorV2(db *mongov2.Storage) *MongoCollector {
	return &MongoCollector{db: db}
}

func (c *MongoCollector) Stat(ctx context.Context) Stat {
	stat, err := c.db.Stats(ctx)
	return NewStat("mongo", stat, err)
}
