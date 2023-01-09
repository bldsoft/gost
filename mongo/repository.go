package mongo

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/bldsoft/gost/repository"
	"github.com/bldsoft/gost/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserEntryCtxKey is the context.Context key to store the user entry. It's used for setting UpdateUserID, CreateUserID fields
var UserEntryCtxKey interface{} = "UserEntry"

type SessionContext = mongo.SessionContext

type IEntityTimeStamp interface {
	SetUpdateFields(cupdateTime time.Time, updateUserID interface{})
	SetCreateFields(createTime time.Time, createUserID interface{})
}

type Repository[T any, U repository.IEntityIDPtr[T]] struct {
	dbcollection   *mongo.Collection
	db             *Storage
	collectionName string
}

func NewRepository[T any, U repository.IEntityIDPtr[T]](db *Storage, collectionName string) *Repository[T, U] {
	return &Repository[T, U]{db: db, collectionName: collectionName}
}

func (r *Repository[T, U]) Name() string {
	return r.collectionName
}

func (r *Repository[T, U]) Collection() *mongo.Collection {
	if r.dbcollection == nil && r.db.IsReady() { // TODO: remove IsReady
		return r.db.Db.Collection(r.collectionName)
	}
	return r.dbcollection
}

func (r *Repository[T, U]) WithTransaction(ctx context.Context, f func(ctx mongo.SessionContext) (interface{}, error)) (interface{}, error) {
	session, err := r.db.Client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	return session.WithTransaction(ctx, f)
}

func (r *Repository[T, U]) projection(opt ...*repository.QueryOptions) interface{} {
	if len(opt) != 0 && len(opt[0].Fields) > 0 {
		var projection bson.D
		for _, field := range opt[0].Fields {
			projection = append(projection, bson.E{Key: field, Value: 1})
		}
		return projection
	}

	return nil
}

func (r *Repository[T, U]) FindOne(ctx context.Context, filter interface{}, opt ...*repository.QueryOptions) (U, error) {
	var result T
	findOneOpt := options.FindOne().SetProjection(r.projection(opt...))
	err := r.Collection().FindOne(ctx, r.where(filter, opt...), findOneOpt).Decode(&result)
	if err != nil && err == mongo.ErrNoDocuments {
		return nil, utils.ErrObjectNotFound
	}
	return &result, err
}

func (r *Repository[T, U]) FindByID(ctx context.Context, id interface{}, options ...*repository.QueryOptions) (U, error) {
	return r.FindOne(ctx, bson.M{"_id": repository.ToRawID[T, U](id)}, options...)
}

func (r *Repository[T, U]) FindByStringIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	return r.findByRawIDs(ctx, repository.StringsToRawIDs[T, U](ids), preserveOrder, options...)
}

func (r *Repository[T, U]) FindByIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
	return r.findByRawIDs(ctx, repository.ToRawIDs[T, U](ids), preserveOrder, options...)
}

func (r *Repository[T, U]) findByRawIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*repository.QueryOptions) ([]U, error) {
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

func (r *Repository[T, U]) Find(ctx context.Context, filter interface{}, opt ...*repository.QueryOptions) ([]U, error) {
	findOpt := options.Find().SetProjection(r.projection(opt...))
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

func (r *Repository[T, U]) GetAll(ctx context.Context, options ...*repository.QueryOptions) ([]U, error) {
	return r.Find(ctx, bson.M{}, options...)
}

func (r *Repository[T, U]) prepareInsertEntity(ctx context.Context, entity U) {
	if entity.IsZeroID() {
		entity.GenerateID()
	}
	r.fillTimeStamp(ctx, entity, true)
}

func (r *Repository[T, U]) Insert(ctx context.Context, entity U) error {
	r.prepareInsertEntity(ctx, entity)
	_, err := r.Collection().InsertOne(ctx, entity)
	return err
}

func (r *Repository[T, U]) InsertMany(ctx context.Context, entities []U) error {
	if len(entities) == 0 {
		return nil
	}
	docs := make([]interface{}, 0, len(entities))
	for _, entity := range entities {
		r.prepareInsertEntity(ctx, entity)
		docs = append(docs, entity)
	}
	_, err := r.Collection().InsertMany(ctx, docs)
	return err
}

func (r *Repository[T, U]) Update(ctx context.Context, entity U, options ...*repository.QueryOptions) error {
	r.fillTimeStamp(ctx, entity, false)
	result, err := r.Collection().ReplaceOne(ctx, r.where(bson.M{"_id": entity.RawID()}, options...), entity)
	if err == nil && result.MatchedCount == 0 {
		return utils.ErrObjectNotFound
	}
	return err
}

func (r *Repository[T, U]) UpdateMany(ctx context.Context, entities []U) error {
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

func (r *Repository[T, U]) UpdateOne(ctx context.Context, filter interface{}, update interface{}, options ...*repository.QueryOptions) error {
	result, err := r.Collection().UpdateOne(ctx, r.where(filter, options...), update)
	if err == nil && result.MatchedCount == 0 {
		return utils.ErrObjectNotFound
	}
	return err
}

func (r *Repository[T, U]) UpdateAndGetByID(ctx context.Context, updateEntity U, returnNewDocument bool, queryOpt ...*repository.QueryOptions) (U, error) {
	opt := options.FindOneAndUpdate()
	if returnNewDocument {
		opt.SetReturnDocument(options.After)
	} else {
		opt.SetReturnDocument(options.Before)
	}
	r.fillTimeStamp(ctx, updateEntity, false)
	res := r.Collection().FindOneAndUpdate(ctx, r.where(bson.M{"_id": updateEntity.RawID()}, queryOpt...), bson.M{"$set": updateEntity}, opt)
	switch {
	case res.Err() == mongo.ErrNoDocuments:
		return nil, utils.ErrObjectNotFound
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

func (r *Repository[T, U]) UpsertOne(ctx context.Context, filter interface{}, update U) error {
	opts := options.Update().SetUpsert(true)
	result, err := r.Collection().UpdateOne(ctx, filter, update, opts)
	if err == nil && result.MatchedCount+result.UpsertedCount == 0 {
		return utils.ErrObjectNotFound
	}
	return err
}

// Delete removes object by id
func (r *Repository[T, U]) Delete(ctx context.Context, id interface{}, options ...*repository.QueryOptions) error {
	id = repository.ToRawID[T, U](id)

	if options != nil {
		if !options[0].Archived {
			_, err := r.Collection().DeleteOne(ctx, bson.M{"_id": id})
			return err
		}
	}

	return r.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{BsonFieldNameArchived: true}})
}

// Delete removes objects
func (r *Repository[T, U]) DeleteMany(ctx context.Context, filter interface{}, options ...*repository.QueryOptions) error {
	if options != nil {
		if !options[0].Archived {
			_, err := r.Collection().DeleteMany(ctx, filter)
			return err
		}
	}
	_, err := r.dbcollection.UpdateMany(ctx, filter, bson.M{"$set": bson.M{BsonFieldNameArchived: true}})
	return err
}

func (r *Repository[T, U]) fillTimeStamp(ctx context.Context, e repository.IEntityID, fillCreateTime bool) {
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
			entityTimestamp.SetCreateFields(*entity.CreateTime, entity.CreateUserID)
		}
	}
}

func (r *Repository[T, U]) where(filter interface{}, options ...*repository.QueryOptions) interface{} {
	if options != nil {
		switch filter := filter.(type) {
		case bson.M:
			var cond interface{}
			var field string

			if !options[0].Archived {
				field = "$or"
				cond = []interface{}{
					bson.M{BsonFieldNameArchived: bson.M{"$exists": false}},
					bson.M{BsonFieldNameArchived: false},
				}
			}

			if _, ok := filter[field]; !ok {
				filter[field] = cond
			}

			recursiveParse(filter, options[0].Filter, "")
		default:
		}
	}

	return filter
}

func (r *Repository[T, U]) AggregateOne(ctx context.Context, pipeline mongo.Pipeline, entity interface{}) error {
	cursor, err := r.Collection().Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return utils.ErrObjectNotFound
	}
	return cursor.Decode(entity)
}
