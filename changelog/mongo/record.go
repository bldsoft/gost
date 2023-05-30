package mongo

import (
	"context"

	"github.com/bldsoft/gost/changelog"
	"github.com/bldsoft/gost/mongo"
)

type Record struct {
	mongo.EntityID    `bson:",inline"`
	*changelog.Record `bson:",inline"`
}

func NewRecord(ctx context.Context, collectionName string, op changelog.Operation) (*Record, error) {
	baseRecord, err := changelog.NewRecord(ctx, collectionName, op, nil)
	if err != nil {
		return nil, err
	}
	rec := &Record{Record: baseRecord}
	rec.GenerateID()
	return rec, nil
}
