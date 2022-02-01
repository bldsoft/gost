package changelog

import (
	"context"

	"github.com/bldsoft/gost/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoRecord struct {
	mongo.EntityID `bson:",inline"`
	*Record        `bson:",inline"`
}

type ChangeLogMongoRepository struct {
	rep *mongo.Repository
}

func NewChangeLogMongoRepository(db *mongo.MongoDb) *ChangeLogMongoRepository {
	return &ChangeLogMongoRepository{mongo.NewRepository(db, "change_log")}
}

func (r *ChangeLogMongoRepository) Insert(ctx context.Context, record *Record) error {
	return r.rep.Insert(ctx, &mongoRecord{Record: record})
}

func (r *ChangeLogMongoRepository) GetRecords(ctx context.Context, filter *Filter) ([]*Record, error) {
	queryFilter := bson.M{}
	if len(filter.Collections) > 0 {
		queryFilter[BsonFieldNameEntity] = bson.M{"$in": filter.Collections}
	}

	timestampFilter := bson.M{}
	if filter.StartTime != 0 {
		timestampFilter["$gte"] = filter.StartTime
	}
	if filter.EndTime != 0 {
		timestampFilter["$lt"] = filter.EndTime
	}
	if len(timestampFilter) > 0 {
		queryFilter[BsonFieldNameTimestamp] = timestampFilter
	}

	if len(filter.EntityID) > 0 {
		var item mongo.EntityID
		if err := item.SetIDFromString(filter.EntityID); err != nil {
			return nil, err
		}
		queryFilter[BsonFieldNameEntityID] = item.GetID()
	}

	opt := &options.FindOptions{}
	cursor, err := r.rep.Collection().Find(ctx, queryFilter, opt.SetSort(bson.M{BsonFieldNameTimestamp: 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	records := make([]*Record, 0)
	return records, cursor.All(ctx, &records)
}

// Compile time checks to ensure your type satisfies an interface
var _ IChangeLogRepository = (*ChangeLogMongoRepository)(nil)
