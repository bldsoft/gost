package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/bldsoft/gost/changelog"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/utils"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	driver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type idParse = func(stringID string) (ID interface{}, err error)

var collectionToType = make(map[string]idParse)

func registerIdParser(collection string, f idParse) {
	collectionToType[collection] = f
}

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
		if len(filter.Collections) != 1 {
			return nil, errors.Errorf("collection cannot be selected unambiguously")
		}
		stringToID, ok := collectionToType[filter.Collections[0]]
		if !ok {
			return nil, utils.ErrObjectNotFound
		}
		id, err := stringToID(filter.EntityID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse id: %w", err)
		}
		queryFilter[changelog.BsonFieldNameEntityID] = id
		queryFilter[changelog.BsonFieldNameEntity] = filter.Collections[0]
	} else if len(filter.Collections) > 0 {
		queryFilter[changelog.BsonFieldNameEntity] = bson.M{"$in": filter.Collections}
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
