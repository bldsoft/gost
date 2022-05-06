package stat

import (
	"context"

	"github.com/bldsoft/gost/mongo"
)

type MongoCollector struct {
	db *mongo.MongoDb
}

func NewMongoCollector(db *mongo.MongoDb) *MongoCollector {
	return &MongoCollector{db: db}
}

func (c *MongoCollector) Stat(ctx context.Context) Stat {
	stat, err := c.db.Stats(ctx)
	return NewStat("mongo", stat, err)
}
