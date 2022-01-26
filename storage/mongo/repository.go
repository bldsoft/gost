package mongo

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/bldsoft/gost/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserEntryCtxKey is the context.Context key to store the user entry. It's used for setting UpdateUserID, CreateUserID fields
var UserEntryCtxKey interface{} = "UserEntry"

type IEntityID interface {
	SetIDFromString(string) error
	GetID() interface{}
	GenerateID()
}

type IEntityTimeStamp interface {
	SetUpdateFields(cupdateTime time.Time, updateUserID interface{})
	SetCreateFields(createTime time.Time, createUserID interface{})
}

type Repository struct {
	dbcollection   *mongo.Collection
	db             *MongoDb
	collectionName string
}

func NewRepository(db *MongoDb, collectionName string) *Repository {
	return &Repository{db: db, collectionName: collectionName}
}

func (r *Repository) Collection() *mongo.Collection {
	if r.dbcollection == nil && r.db.IsReady() {
		r.dbcollection = r.db.Db.Collection(r.collectionName)
	}
	return r.dbcollection
}

func (r *Repository) FindOne(ctx context.Context, filter interface{}, result interface{}, options ...*QueryOptions) error {
	err := r.Collection().FindOne(ctx, r.where(filter, options...)).Decode(result)
	if err != nil && err == mongo.ErrNoDocuments {
		return utils.ErrObjectNotFound
	}
	return err
}

func (r *Repository) FindByID(ctx context.Context, id string, result interface{}, options ...*QueryOptions) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	return r.FindOne(ctx, bson.M{"_id": objID}, result, options...)
}

func (r *Repository) Find(ctx context.Context, filter interface{}, results interface{}, options ...*QueryOptions) error {
	cur, err := r.Collection().Find(ctx, r.where(filter, options...))
	if err != nil {
		return err
	}
	return cur.All(ctx, results)
}

func (r *Repository) GetAll(ctx context.Context, results interface{}, options ...*QueryOptions) {
	r.Find(ctx, bson.M{}, results, options...)
}

func (r *Repository) Insert(ctx context.Context, entity IEntityID) error {
	entity.GenerateID()
	r.fillTimeStamp(ctx, entity, true)
	_, err := r.Collection().InsertOne(ctx, entity)
	return err
}

func (r *Repository) Update(ctx context.Context, entity IEntityID, options ...*QueryOptions) error {
	r.fillTimeStamp(ctx, entity, false)
	result, err := r.Collection().ReplaceOne(ctx, r.where(bson.M{"_id": entity.GetID()}, options...), entity)
	if err == nil && result.MatchedCount == 0 {
		return utils.ErrObjectNotFound
	}
	return err
}

func (r *Repository) UpdateMany(ctx context.Context, entities interface{}) error {
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

func (r *Repository) UpdateOne(ctx context.Context, filter interface{}, update interface{}, options ...*QueryOptions) error {
	result, err := r.Collection().UpdateOne(ctx, r.where(filter, options...), update)
	if err == nil && result.MatchedCount == 0 {
		return utils.ErrObjectNotFound
	}
	return err
}

func (r *Repository) UpdateAndGetByID(ctx context.Context, updateEntity IEntityID, result interface{}, queryOpt ...*QueryOptions) error {
	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)
	r.fillTimeStamp(ctx, updateEntity, false)
	res := r.Collection().FindOneAndUpdate(ctx, r.where(bson.M{"_id": updateEntity.GetID()}, queryOpt...), bson.M{"$set": updateEntity}, opt)
	switch {
	case res.Err() == mongo.ErrNoDocuments:
		return utils.ErrObjectNotFound
	case res.Err() != nil:
		return res.Err()
	default:
		return res.Decode(result)
	}
}

func (r *Repository) UpsertOne(ctx context.Context, filter interface{}, update interface{}) error {
	opts := options.Update().SetUpsert(true)
	result, err := r.Collection().UpdateOne(ctx, filter, update, opts)
	if err == nil && result.MatchedCount+result.UpsertedCount == 0 {
		return utils.ErrObjectNotFound
	}
	return err
}

//Delete removes object by id
func (r *Repository) Delete(ctx context.Context, e IEntityID, options ...*QueryOptions) error {
	if options != nil {
		if !options[0].Archived {
			_, err := r.Collection().DeleteOne(ctx, bson.M{"_id": e.GetID()})
			return err
		}
	}

	return r.UpdateOne(ctx, bson.M{"_id": e.GetID()}, bson.M{"$set": bson.M{bsonFieldNameArchived: true}})
}

func (r *Repository) fillTimeStamp(ctx context.Context, e IEntityID, fillCreateTime bool) {
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

func (r *Repository) where(filter interface{}, options ...*QueryOptions) interface{} {
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

func (r *Repository) AggregateOne(ctx context.Context, pipeline mongo.Pipeline, entity interface{}) error {
	cursor, err := r.Collection().Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return errors.New("Not found")
	}
	return cursor.Decode(entity)
}
