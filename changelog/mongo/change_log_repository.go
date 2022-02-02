package mongo

import (
	"context"

	"github.com/bldsoft/gost/changelog"
	"github.com/bldsoft/gost/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoRecord struct {
	mongo.EntityID    `bson:",inline"`
	*changelog.Record `bson:",inline"`
}

type ChangeLogRepository struct {
	rep *mongo.Repository[*mongoRecord]
}

func NewChangeLogRepository(db *mongo.MongoDb) *ChangeLogRepository {
	return &ChangeLogRepository{mongo.NewRepository[*mongoRecord](db, "change_log")}
}

func (r *ChangeLogRepository) Insert(ctx context.Context, record *changelog.Record) error {
	return r.rep.Insert(ctx, &mongoRecord{Record: record})
}

func (r *ChangeLogRepository) GetRecords(ctx context.Context, filter *changelog.Filter) ([]*changelog.Record, error) {
	queryFilter := bson.M{}
	if len(filter.Collections) > 0 {
		queryFilter[changelog.BsonFieldNameEntity] = bson.M{"$in": filter.Collections}
	}

	timestampFilter := bson.M{}
	if filter.StartTime != 0 {
		timestampFilter["$gte"] = filter.StartTime
	}
	if filter.EndTime != 0 {
		timestampFilter["$lt"] = filter.EndTime
	}
	if len(timestampFilter) > 0 {
		queryFilter[changelog.BsonFieldNameTimestamp] = timestampFilter
	}

	if len(filter.EntityID) > 0 {
		var item mongo.EntityID
		if err := item.SetIDFromString(filter.EntityID); err != nil {
			return nil, err
		}
		queryFilter[changelog.BsonFieldNameEntityID] = item.GetID()
	}

	opt := &options.FindOptions{}
	cursor, err := r.rep.Collection().Find(ctx, queryFilter, opt.SetSort(bson.M{changelog.BsonFieldNameTimestamp: 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	records := make([]*changelog.Record, 0)
	return records, cursor.All(ctx, &records)
}

// Compile time checks to ensure your type satisfies an interface
var _ changelog.IChangeLogRepository = (*ChangeLogRepository)(nil)
