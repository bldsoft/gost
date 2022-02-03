package mongo

import (
	"context"

	"github.com/bldsoft/gost/changelog"
	"github.com/bldsoft/gost/mongo"
)

type record struct {
	mongo.EntityID    `bson:",inline"`
	*changelog.Record `bson:",inline"`
}

func newRecord(ctx context.Context, collectionName string, op changelog.Operation, entity changelog.EntityID) (*record, error) {
	rec, err := changelog.NewRecord(ctx, collectionName, op, entity)
	if err != nil {
		return nil, err
	}
	return &record{Record: rec}, nil
}
