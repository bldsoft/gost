package mongo

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	driver "go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/bldsoft/gost/changelog"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
)

// NameSortJoin describes the collection that holds the display name of the changelog user.
// When set, sorting by changelog.SortFieldUser orders records by that name instead of the user ID.
type NameSortJoin struct {
	From         string // collection to join, e.g. "user"
	LocalField   string // field on change_log, e.g. "userID"
	ForeignField string // field on the joined collection, e.g. "_id"
	NameField    string // field to sort by, e.g. "name"
}

func (j NameSortJoin) configured() bool {
	return j != NameSortJoin{}
}

const (
	nameSortJoinAs    = "_sortUser"
	nameSortNameField = "_sortUserName"
)

type ChangeLogRepository struct {
	rep          mongo.Repository[Record, *Record]
	nameSortJoin NameSortJoin
}

func NewChangeLogRepository(db *mongo.Storage) *ChangeLogRepository {
	r := &ChangeLogRepository{rep: mongo.NewRepository[Record](db, "change_log")}

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

// SetNameSortJoin enables sorting by user name for changelog.SortFieldUser. The zero value disables it.
func (r *ChangeLogRepository) SetNameSortJoin(join NameSortJoin) {
	r.nameSortJoin = join
}

func (r *ChangeLogRepository) Insert(ctx context.Context, record *Record) error {
	_, err := r.rep.Collection().InsertOne(ctx, record)
	return err
}

func (r *ChangeLogRepository) InsertMany(ctx context.Context, records []*Record) error {
	_, err := r.rep.Collection().InsertMany(ctx, records)
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

func (r *ChangeLogRepository) GetRecords(ctx context.Context, params *changelog.RecordsParams) (*changelog.Records, error) {
	filter, err := r.recordsFilter(params.Filter)
	if err != nil {
		return nil, err
	}

	var cursor *driver.Cursor
	if params.Sort.Field == changelog.SortFieldUser && r.nameSortJoin.configured() {
		cursor, err = r.rep.Collection().Aggregate(ctx, r.nameSortPipeline(filter, params),
			options.Aggregate().SetCollation(&options.Collation{Locale: "en", Strength: 2}))
	} else {
		cursor, err = r.rep.Collection().Find(ctx, filter, options.Find().
			SetSort(r.recordsSort(params.Sort)).
			SetSkip(params.Offset).
			SetLimit(params.Limit))
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cursor.Close(ctx)
	}()

	var res changelog.Records
	if err := cursor.All(ctx, &res.Records); err != nil {

		return nil, err
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

	if filter.Details != nil {
		filter.Details.Filter(queryFilter)
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

// nameSortPipeline must run with a case-insensitive collation.
func (r *ChangeLogRepository) nameSortPipeline(filter bson.M, params *changelog.RecordsParams) driver.Pipeline {
	join := r.nameSortJoin
	order := 1
	if params.Sort.Order == repository.SortOrderDESC {
		order = -1
	}

	pipeline := driver.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: join.From},
			{Key: "localField", Value: join.LocalField},
			{Key: "foreignField", Value: join.ForeignField},
			{Key: "as", Value: nameSortJoinAs},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: nameSortNameField, Value: bson.D{{Key: "$ifNull", Value: bson.A{
				bson.D{{Key: "$arrayElemAt", Value: bson.A{"$" + nameSortJoinAs + "." + join.NameField, 0}}},
				"",
			}}}},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: nameSortNameField, Value: order},
			{Key: changelog.BsonFieldNameTimestamp, Value: -1},
			{Key: "_id", Value: -1},
		}}},
	}
	if params.Offset > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: params.Offset}})
	}
	if params.Limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: params.Limit}})
	}

	return append(pipeline, bson.D{{Key: "$project", Value: bson.D{
		{Key: nameSortJoinAs, Value: 0},
		{Key: nameSortNameField, Value: 0},
	}}})
}

// Compile time checks to ensure your type satisfies an interface
var _ changelog.IChangeLogRepository = (*ChangeLogRepository)(nil)
