package healthcheck

import (
	"context"

	"github.com/bldsoft/gost/mongo"
)

type MongoHealthChecker struct {
	db *mongo.MongoDb
}

func NewMongoHealthChecker(db *mongo.MongoDb) *MongoHealthChecker {
	return &MongoHealthChecker{db: db}
}

func (c *MongoHealthChecker) CheckHealth(ctx context.Context) Health {
	stat, err := c.db.Stats(ctx)
	return NewHealth("mongo", stat, err)
}
