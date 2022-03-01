package mongo

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/bldsoft/gost/repository"
	"github.com/bldsoft/gost/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserEntryCtxKey is the context.Context key to store the user entry. It's used for setting UpdateUserID, CreateUserID fields
var UserEntryCtxKey interface{} = "UserEntry"

type SessionContext = mongo.SessionContext

type IEntityID interface {
	SetIDFromString(string) error
	GetID() interface{}
	GenerateID()
}

type IEntityTimeStamp interface {
	SetUpdateFields(cupdateTime time.Time, updateUserID interface{})
	SetCreateFields(createTime time.Time, createUserID interface{})
}

type IEntityIDPtr[T any] interface {
	*T
	IEntityID
}

type Repository[T any, U IEntityIDPtr[T]] struct {
	dbcollection   *mongo.Collection
	db             *MongoDb
	collectionName string
}

func NewRepository[T any, U IEntityIDPtr[T]](db *MongoDb, collectionName string) *Repository[T, U] {
	return &Repository[T, U]{db: db, collectionName: collectionName}
}

func (r *Repository[T, U]) Name() string {
	return r.collectionName
}

func (r *Repository[T, U]) Collection() *mongo.Collection {
	if r.dbcollection == nil && r.db.IsReady() {
		r.dbcollection = r.db.Db.Collection(r.collectionName)
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

func (r *Repository[T, U]) FindOne(ctx context.Context, filter interface{}, options ...*repository.QueryOptions) (U, error) {
	var result T
	err := r.Collection().FindOne(ctx, r.where(filter, options...)).Decode(&result)
	if err != nil && err == mongo.ErrNoDocuments {
		return nil, utils.ErrObjectNotFound
	}
	return &result, err
}

func (r *Repository[T, U]) FindByID(ctx context.Context, id interface{}, options ...*repository.QueryOptions) (U, error) {
	var err error
	switch v := id.(type) {
	case string:
		id, err = primitive.ObjectIDFromHex(v)
	case IEntityID:
		id = v.GetID()
	}
	if err != nil {
		return nil, err
	}
	return r.FindOne(ctx, bson.M{"_id": id}, options...)
}

func (r *Repository[T, U]) Find(ctx context.Context, filter interface{}, options ...*repository.QueryOptions) ([]U, error) {
	cur, err := r.Collection().Find(ctx, r.where(filter, options...))
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

func (r *Repository[T, U]) Insert(ctx context.Context, entity U) error {
	entity.GenerateID()
	r.fillTimeStamp(ctx, entity, true)
	_, err := r.Collection().InsertOne(ctx, entity)
	return err
}

func (r *Repository[T, U]) Update(ctx context.Context, entity U, options ...*repository.QueryOptions) error {
	r.fillTimeStamp(ctx, entity, false)
	result, err := r.Collection().ReplaceOne(ctx, r.where(bson.M{"_id": entity.GetID()}, options...), entity)
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
			entity := s.Index(i).Interface().(IEntityID)
			r.fillTimeStamp(ctx, entity, false)
			opertaion := mongo.NewReplaceOneModel()
			opertaion.SetFilter(bson.M{"_id": entity.GetID()})
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
	res := r.Collection().FindOneAndUpdate(ctx, r.where(bson.M{"_id": updateEntity.GetID()}, queryOpt...), bson.M{"$set": updateEntity}, opt)
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

//Delete removes object by id
func (r *Repository[T, U]) Delete(ctx context.Context, id interface{}, options ...*repository.QueryOptions) error {
	var err error
	switch v := id.(type) {
	case string:
		id, err = primitive.ObjectIDFromHex(v)
	case IEntityID:
		id = v.GetID()
	}
	if err != nil {
		return err
	}

	if options != nil {
		if !options[0].Archived {
			_, err := r.Collection().DeleteOne(ctx, bson.M{"_id": id})
			return err
		}
	}

	return r.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{bsonFieldNameArchived: true}})
}

func (r *Repository[T, U]) fillTimeStamp(ctx context.Context, e IEntityID, fillCreateTime bool) {
	if entityTimestamp, ok := e.(IEntityTimeStamp); ok {
		user, ok := ctx.Value(UserEntryCtxKey).(IEntityID)
		if !ok {
			return
		}

		now := time.Now().UTC()
		entityTimestamp.SetUpdateFields(now, user.GetID())
		if fillCreateTime {
			entityTimestamp.SetCreateFields(now, user.GetID())
		} else {
			projection := bson.D{
				{Key: BsonFieldNameCreateUserID, Value: 1},
				{Key: BsonFieldNameCreateTime, Value: 1},
			}

			result := r.Collection().FindOne(ctx,
				bson.M{"_id": e.GetID()},
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
					bson.M{bsonFieldNameArchived: bson.M{"$exists": false}},
					bson.M{bsonFieldNameArchived: false},
				}
			}

			if _, ok := filter[field]; !ok {
				filter[field] = cond
			}
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
