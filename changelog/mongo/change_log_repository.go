package mongo

import (
	"context"
	"time"

	"github.com/bldsoft/gost/changelog"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	driver "go.mongodb.org/mongo-driver/mongo"
)

type ChangeLogRepository struct {
	rep mongo.Repository[record, *record]
}

func NewChangeLogRepository(db *mongo.Storage) *ChangeLogRepository {
	r := &ChangeLogRepository{mongo.NewRepository[record, *record](db, "change_log")}

	indexes := []driver.IndexModel{
		{Keys: bson.D{bson.E{Key: changelog.BsonFieldNameUserID, Value: 1}}},
		{Keys: bson.D{bson.E{Key: changelog.BsonFieldNameEntityID, Value: 1}}},
		{Keys: bson.D{bson.E{Key: changelog.BsonFieldNameData, Value: "text"}}},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.rep.Collection().Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.ErrorWithFields(log.Fields{"err": err}, "Failed to create indexes for change_log")
	}

	return r
}

func (r *ChangeLogRepository) Insert(ctx context.Context, record *record) error {
	_, err := r.rep.Collection().InsertOne(ctx, record)
	return err
}

func (r *ChangeLogRepository) FindByID(ctx context.Context, id string, options ...*repository.QueryOptions) (*changelog.Record, error) {
	record, err := r.rep.FindByID(ctx, id, options...)
	if err != nil {
		return nil, err
	}
	return record.Record, nil
}

func (r *ChangeLogRepository) FindByIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) (res []*changelog.Record, err error) {
	records, err := r.rep.FindByStringIDs(ctx, ids, preserveOrder, options...)
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		res = append(res, record.Record)
	}
	return res, nil
}

func (r *ChangeLogRepository) GetRecords(ctx context.Context, params *changelog.RecordsParams, options ...*repository.QueryOptions) (*changelog.Records, error) {
	filter, err := r.recordsFilter(params.Filter)
	if err != nil {
		return nil, err
	}

	var res changelog.Records

	if len(options) == 0 {
		options = append(options, &repository.QueryOptions{})
	}
	options[0].Sort = r.recordsSort(params.Sort)
	options[0].Skip = params.Offset
	options[0].Limit = params.Limit

	records, err := r.rep.Find(ctx, filter, options...)
	if err != nil {
		return nil, err
	}
	if len(records) != 0 {
		res.Records = make([]changelog.Record, len(records))
		for i, r := range records {
			res.Records[i] = *r.Record
		}
	}

	res.TotalCount, err = r.rep.Collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ChangeLogRepository) recordsFilter(filter *changelog.Filter) (bson.M, error) {
	queryFilter := make(bson.M)
	if filter == nil {
		return queryFilter, nil
	}
	if len(filter.EntityID) > 0 {
		if len(filter.Entities) != 1 {
			return nil, errors.Errorf("unambiguous collection")
		}
		queryFilter[changelog.BsonFieldNameEntityID] = filter.EntityID
		queryFilter[changelog.BsonFieldNameEntity] = filter.Entities[0]
	} else if len(filter.Entities) > 0 {
		queryFilter[changelog.BsonFieldNameEntity] = bson.M{"$in": filter.Entities}
	}

	if len(filter.UserIDs) > 0 {
		queryFilter[changelog.BsonFieldNameUserID] = bson.M{"$in": filter.UserIDs}
	}

	if len(filter.Operations) > 0 {
		queryFilter[changelog.BsonFieldNameOperation] = bson.M{"$in": filter.Operations}
	}

	if filter.Search != nil && len(*filter.Search) > 0 {
		queryFilter["$text"] = bson.D{{Key: "$search", Value: *filter.Search}}
	}

	timestampFilter := bson.M{}
	if filter.From != nil {
		timestampFilter["$gte"] = *filter.From
	}
	if filter.To != nil {
		timestampFilter["$lt"] = *filter.To
	}
	if len(timestampFilter) > 0 {
		queryFilter[changelog.BsonFieldNameTimestamp] = timestampFilter
	}

	return queryFilter, nil
}

func (r *ChangeLogRepository) recordsSort(sort changelog.Sort) bson.D {
	fieldName := changelog.BsonFieldNameTimestamp
	switch sort.Field {
	case changelog.SortFieldTimestamp:
		fieldName = changelog.BsonFieldNameTimestamp
	case changelog.SortFieldUser:
		fieldName = changelog.BsonFieldNameUserID
	case changelog.SortFieldOperation:
		fieldName = changelog.BsonFieldNameOperation
	case changelog.SortFieldEntity:
		fieldName = changelog.BsonFieldNameEntity
	}
	order := 1
	if sort.Order == repository.SortOrderDESC {
		order = -1
	}
	return bson.D{{Key: fieldName, Value: order}}
}

// Compile time checks to ensure your type satisfies an interface
var _ changelog.IChangeLogRepository = (*ChangeLogRepository)(nil)
