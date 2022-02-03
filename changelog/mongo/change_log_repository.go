package mongo

import (
	"context"
	"time"

	"github.com/bldsoft/gost/changelog"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"go.mongodb.org/mongo-driver/bson"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChangeLogRepository struct {
	rep *mongo.Repository[record]
}

func NewChangeLogRepository(db *mongo.MongoDb) *ChangeLogRepository {
	r := &ChangeLogRepository{mongo.NewRepository[record](db, "change_log")}
	db.AddOnConnectHandler(func() {
		indexes := []driver.IndexModel{
			{Keys: bson.D{bson.E{Key: changelog.BsonFieldNameUserID, Value: 1}}},
			{Keys: bson.D{bson.E{Key: changelog.BsonFieldNameEntityID, Value: 1}}},
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := r.rep.Collection().Indexes().CreateMany(ctx, indexes)
		if err != nil {
			log.ErrorWithFields(log.Fields{"err": err}, "Failed to create indexes for change_log")
		}
	})
	return r
}

func (r *ChangeLogRepository) Insert(ctx context.Context, record *record) error {
	_, err := r.rep.Collection().InsertOne(ctx, record)
	return err
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
