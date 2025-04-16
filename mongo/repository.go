package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	defaultOpt = repository.QueryOptions{
		Archived: false,
	}
)

// UserEntryCtxKey is the context.Context key to store the user entry. It's used for setting UpdateUserID, CreateUserID fields
var UserEntryCtxKey interface{} = "UserEntry"

type SessionContext = mongo.SessionContext

type IEntityTimeStamp interface {
	SetUpdateFields(cupdateTime time.Time, updateUserID interface{})
	SetCreateFields(createTime time.Time, createUserID interface{})
}

type BaseRepository[T any, U repository.IEntityIDPtr[T]] struct {
	dbcollection   *mongo.Collection
	db             *Storage
	collectionName string
}

func NewRepository[T any, U repository.IEntityIDPtr[T]](db *Storage, collectionName string) Repository[T, U] {
	return &BaseRepository[T, U]{db: db, collectionName: collectionName}
}

func (r *BaseRepository[T, U]) Name() string {
	return r.collectionName
}

func (r *BaseRepository[T, U]) Collection() *mongo.Collection {
	if r.dbcollection == nil && r.db.IsReady() { // TODO: remove IsReady
		return r.db.Db.Collection(r.collectionName)
	}
	return r.dbcollection
}

func (r *BaseRepository[T, U]) WithTransaction(ctx context.Context, f func(ctx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	session, err := r.db.Client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	return session.WithTransaction(ctx, f)
}

func (r *BaseRepository[T, U]) projection(opt ...*repository.QueryOptions) interface{} {
	if len(opt) != 0 && len(opt[0].Fields) > 0 {
		var projection bson.D
		set := make(map[string]struct{})
		for _, field := range opt[0].Fields {
			if _, ok := set[field]; ok {
				continue
			}
			projection = append(projection, bson.E{Key: field, Value: 1})
			set[field] = struct{}{}
		}
		return projection
	}

	return nil
}

func (r *BaseRepository[T, U]) FindOne(ctx context.Context, filter interface{}, opt ...*repository.QueryOptions) (U, error) {
	var result T
	findOneOpt := options.FindOne().SetProjection(r.projection(opt...))
	err := r.Collection().FindOne(ctx, r.where(filter, opt...), findOneOpt).Decode(&result)
	if err != nil && err == mongo.ErrNoDocuments {
		return nil, repository.ErrNotFound
	}
	return &result, err
}

func (r *BaseRepository[T, U]) FindByID(ctx context.Context, id interface{}, options ...*repository.QueryOptions) (U, error) {
	return r.FindOne(ctx, bson.M{"_id": repository.ToRawID[T, U](id)}, options...)
}

func (r *BaseRepository[T, U]) FindByStringIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	return r.findByRawIDs(ctx, repository.StringsToRawIDs[T, U](ids), preserveOrder, options...)
}

func (r *BaseRepository[T, U]) FindByIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	return r.findByRawIDs(ctx, repository.ToRawIDs[T, U](ids), preserveOrder, options...)
}

func (r *BaseRepository[T, U]) findByRawIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	entities, err := r.Find(ctx, bson.M{"_id": bson.M{"$in": ids}}, options...)
	if err != nil {
		return nil, err
	}

	if preserveOrder {
		entityById := make(map[interface{}]U)
		for _, entity := range entities {
			entityById[entity.RawID()] = entity
		}

		result := make([]U, len(ids))
		for i, id := range ids {
			result[i] = entityById[id]
		}
		return result, nil
	}

	return entities, nil
}

func (r *BaseRepository[T, U]) Find(ctx context.Context, filter interface{}, opt ...*repository.QueryOptions) ([]U, error) {
	findOpt := options.Find().
		SetProjection(r.projection(opt...))

	if len(opt) != 0 {
		opt := opt[0]
		if len(opt.Sort) != 0 {
			findOpt = findOpt.SetSort(r.sort(opt.Sort))
		}
		if opt.Limit > 0 {
			findOpt = findOpt.SetLimit(opt.Limit)
		}
		if opt.Offset > 0 {
			findOpt = findOpt.SetSkip(opt.Offset)
		}
	}

	cur, err := r.Collection().Find(ctx, r.where(filter, opt...), findOpt)
	if err != nil {
		return nil, err
	}
	results := make([]U, 0)
	if err = cur.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *BaseRepository[T, U]) Count(ctx context.Context, filter interface{}, opt ...*repository.QueryOptions) (int64, error) {
	return r.Collection().CountDocuments(ctx, r.where(filter, opt...))
}

func (r *BaseRepository[T, U]) sort(opt repository.SortOpt) bson.D {
	var res bson.D
	for _, sortParam := range opt {
		order := 1
		if sortParam.Desc {
			order = -1
		}
		res = append(res, bson.E{Key: sortParam.Field, Value: order})
	}
	return res
}

func (r *BaseRepository[T, U]) GetAll(ctx context.Context, options ...*repository.QueryOptions) ([]U, error) {
	return r.Find(ctx, bson.M{}, options...)
}

func (r *BaseRepository[T, U]) prepareInsertEntity(ctx context.Context, entity U) {
	if entity.IsZeroID() {
		entity.GenerateID()
	}
	r.fillTimeStamp(ctx, entity, true)
}

func (r *BaseRepository[T, U]) Insert(ctx context.Context, entity U, opt ...*repository.ModifyOptions) error {
	if len(opt) > 0 && opt[0].DuplicateFilter != nil {
		if err := r.hasDuplicate(ctx, opt[0].DuplicateFilter); err != nil {
			return err
		}
	}

	r.prepareInsertEntity(ctx, entity)
	_, err := r.Collection().InsertOne(ctx, entity)
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%w: %w", repository.ErrAlreadyExists, err)
	}
	return err
}

func (r *BaseRepository[T, U]) InsertMany(ctx context.Context, entities []U, opt ...*repository.ModifyOptions) error {
	if len(entities) == 0 {
		return nil
	}
	if len(opt) > 0 && opt[0].DuplicateFilter != nil {
		if err := r.hasDuplicate(ctx, opt[0].DuplicateFilter); err != nil {
			return err
		}
	}

	docs := make([]interface{}, 0, len(entities))
	for _, entity := range entities {
		r.prepareInsertEntity(ctx, entity)
		docs = append(docs, entity)
	}
	_, err := r.Collection().InsertMany(ctx, docs)
	return err
}

func (r *BaseRepository[T, U]) Update(ctx context.Context, entity U, options ...*repository.ModifyOptions) error {
	r.fillTimeStamp(ctx, entity, false)
	var qOpt *repository.QueryOptions
	if len(options) > 0 {
		qOpt = &options[0].QueryOptions
		if options[0].DuplicateFilter != nil {
			if err := r.hasDuplicate(ctx, options[0].DuplicateFilter); err != nil {
				return err
			}
		}
	}
	result, err := r.Collection().ReplaceOne(ctx, r.where(bson.M{"_id": entity.RawID()}, qOpt), entity)
	if err == nil && result.MatchedCount == 0 {
		return repository.ErrNotFound
	}
	return err
}

func (r *BaseRepository[T, U]) UpdateMany(ctx context.Context, entities []U) error {
	switch reflect.TypeOf(entities).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(entities)
		size := s.Len()
		operations := make([]mongo.WriteModel, 0, size)
		for i := 0; i < size; i++ {
			entity := s.Index(i).Interface().(repository.IEntityID)
			r.fillTimeStamp(ctx, entity, false)
			opertaion := mongo.NewReplaceOneModel()
			opertaion.SetFilter(bson.M{"_id": entity.RawID()})
			opertaion.SetReplacement(entity)
			operations = append(operations, opertaion)
		}

		bulkOption := options.BulkWriteOptions{}
		result, err := r.Collection().BulkWrite(ctx, operations, &bulkOption)
		if err == nil && result.MatchedCount == 0 {
			return mongo.ErrNoDocuments
		}
		return err
	default:
		return errors.New("entities must be a slice")
	}
}

func (r *BaseRepository[T, U]) UpdateOne(ctx context.Context, filter interface{}, update interface{}, options ...*repository.ModifyOptions) error {
	if len(options) > 0 && options[0].DuplicateFilter != nil {
		if err := r.hasDuplicate(ctx, options[0].DuplicateFilter); err != nil {
			return err
		}
	}

	var qOpt *repository.QueryOptions
	if len(options) > 0 {
		qOpt = &options[0].QueryOptions
	}
	result, err := r.Collection().UpdateOne(ctx, r.where(filter, qOpt), update)
	if err == nil && result.MatchedCount == 0 {
		return repository.ErrNotFound
	}
	return err
}

func (r *BaseRepository[T, U]) UpdateAndGetByID(ctx context.Context, updateEntity U, returnNewDocument bool, opts ...*repository.ModifyOptions) (U, error) {
	opt := options.FindOneAndUpdate()
	if returnNewDocument {
		opt.SetReturnDocument(options.After)
	} else {
		opt.SetReturnDocument(options.Before)
	}
	var qOpt *repository.QueryOptions
	if len(opts) > 0 {
		qOpt = &opts[0].QueryOptions
	}
	r.fillTimeStamp(ctx, updateEntity, false)
	res := r.Collection().FindOneAndUpdate(ctx, r.where(bson.M{"_id": updateEntity.RawID()}, qOpt), bson.M{"$set": updateEntity}, opt)
	switch {
	case res.Err() == mongo.ErrNoDocuments:
		return nil, repository.ErrNotFound
	case res.Err() != nil:
		return nil, res.Err()
	default:
		var result T
		if err := res.Decode(&result); err != nil {
			return nil, err
		}
		return &result, nil
	}
}

func (r *BaseRepository[T, U]) Upsert(ctx context.Context, entity U, opt ...*repository.QueryOptions) error {
	return r.UpsertOne(ctx, r.where(bson.M{"_id": entity.RawID()}, opt...), entity)
}

func (r *BaseRepository[T, U]) UpsertOne(ctx context.Context, filter interface{}, update U) error {
	opts := options.Update().SetUpsert(true)
	result, err := r.Collection().UpdateOne(ctx, filter, bson.M{"$set": update}, opts)
	if err == nil && result.MatchedCount+result.UpsertedCount == 0 {
		return repository.ErrNotFound
	}
	return err
}

// Delete removes object by id
func (r *BaseRepository[T, U]) Delete(ctx context.Context, id interface{}, options ...*repository.QueryOptions) error {
	id = repository.ToRawID[T, U](id)

	if options != nil {
		if !options[0].Archived {
			_, err := r.Collection().DeleteOne(ctx, bson.M{"_id": id})
			return err
		}
	}

	err := r.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{BsonFieldNameDeleteTime: time.Now(), BsonFieldNameArchived: true}})
	return err
}

// Delete removes objects
func (r *BaseRepository[T, U]) DeleteMany(ctx context.Context, filter interface{}, options ...*repository.QueryOptions) error {
	if options != nil {
		if !options[0].Archived {
			_, err := r.Collection().DeleteMany(ctx, filter)
			return err
		}
	}
	_, err := r.Collection().UpdateMany(ctx, filter, bson.M{"$set": bson.M{BsonFieldNameDeleteTime: time.Now(), BsonFieldNameArchived: true}})
	return err
}

func (r *BaseRepository[T, U]) fillTimeStamp(ctx context.Context, e repository.IEntityID, fillCreateTime bool) {
	if entityTimestamp, ok := e.(IEntityTimeStamp); ok {
		user, ok := ctx.Value(UserEntryCtxKey).(repository.IEntityID)
		if !ok {
			return
		}

		now := time.Now().UTC()
		entityTimestamp.SetUpdateFields(now, user.RawID())
		if fillCreateTime {
			entityTimestamp.SetCreateFields(now, user.RawID())
		} else {
			projection := bson.D{
				{Key: BsonFieldNameCreateUserID, Value: 1},
				{Key: BsonFieldNameCreateTime, Value: 1},
			}

			result := r.Collection().FindOne(ctx,
				bson.M{"_id": e.RawID()},
				options.FindOne().SetProjection(projection))

			var entity EntityTimeStamp
			result.Decode(&entity)
			if entity.CreateTime == nil {
				entity.CreateTime = &time.Time{}
			}
			entityTimestamp.SetCreateFields(*entity.CreateTime, entity.CreateUserID)
		}
	}
}

func (r *BaseRepository[T, U]) where(filter interface{}, options ...*repository.QueryOptions) interface{} {
	if len(options) == 0 {
		options = append(options, &defaultOpt)
	}
	switch filter := filter.(type) {
	case bson.M:
		if !options[0].Archived {
			nonArchivedCond := bson.A{
				bson.M{BsonFieldNameArchived: bson.M{"$exists": false}, BsonFieldNameDeleteTime: bson.M{"$exists": false}},
				bson.M{BsonFieldNameArchived: false},
			}

			if cond, ok := filter["$or"]; ok {
				filter["$and"] = bson.A{
					bson.M{"$or": nonArchivedCond},
					bson.M{"$or": cond},
				}
				delete(filter, "$or")
			} else {
				filter["$or"] = nonArchivedCond
			}
		}

		RecursiveParse(filter, options[0].Filter, "")
	default:
	}
	return filter
}

func (r *BaseRepository[T, U]) AggregateOne(ctx context.Context, pipeline mongo.Pipeline, entity interface{}) error {
	cursor, err := r.Collection().Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return repository.ErrNotFound
	}
	return cursor.Decode(entity)
}

func (r *BaseRepository[T, U]) hasDuplicate(ctx context.Context, filter interface{}) error {
	count, err := r.Count(ctx, filter)
	if err != nil {
		return err
	}
	if count > 0 {
		return repository.ErrAlreadyExists
	}

	return nil
}
