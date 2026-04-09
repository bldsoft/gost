package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	defaultOpt = repository.QueryOptions{
		Archived: false,
	}
)

type contextKey string

// UserEntryCtxKey is the context.Context key to store the user entry. It's used for setting UpdateUserID, CreateUserID fields
var UserEntryCtxKey contextKey = "UserEntry"

func WithUserEntry(ctx context.Context, user repository.IEntityID) context.Context {
	return context.WithValue(ctx, UserEntryCtxKey, user)
}

func GetUserEntry(ctx context.Context) (repository.IEntityID, bool) {
	user, ok := ctx.Value(UserEntryCtxKey).(repository.IEntityID)
	return user, ok
}

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
	if r.dbcollection == nil && r.db.IsReady() {
		return r.db.Db.Collection(r.collectionName)
	}
	return r.dbcollection
}

func (r *BaseRepository[T, U]) WithTransaction(ctx context.Context, f func(ctx context.Context) (interface{}, error)) (interface{}, error) {
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
	if errors.Is(err, mongo.ErrNoDocuments) {
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
	if len(ids) == 0 {
		return []U{}, nil
	}

	entities, err := r.Find(ctx, bson.M{"_id": bson.M{"$in": ids}}, options...)
	if err != nil || !preserveOrder {
		return entities, err
	}

	entityById := make(map[any]U, len(entities))
	for _, ent := range entities {
		entityById[ent.RawID()] = ent
	}

	res := make([]U, len(ids))
	for _, id := range ids {
		if ent, ok := entityById[id]; ok {
			res = append(res, ent)
		}
	}

	return res, nil
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

func (r *BaseRepository[T, U]) Insert(ctx context.Context, entity U) error {
	r.prepareInsertEntity(ctx, entity)
	_, err := r.Collection().InsertOne(ctx, entity)
	return wrapErr(err)
}

func (r *BaseRepository[T, U]) InsertMany(ctx context.Context, entities []U) error {
	if len(entities) == 0 {
		return nil
	}
	docs := make([]interface{}, 0, len(entities))
	for _, entity := range entities {
		r.prepareInsertEntity(ctx, entity)
		docs = append(docs, entity)
	}
	_, err := r.Collection().InsertMany(ctx, docs)
	return wrapErr(err)
}

func (r *BaseRepository[T, U]) Update(ctx context.Context, entity U, options ...*repository.QueryOptions) error {
	r.fillTimeStamp(ctx, entity, false)
	result, err := r.Collection().ReplaceOne(ctx, r.where(bson.M{"_id": entity.RawID()}, options...), entity)
	if err != nil {
		return wrapErr(err)
	}
	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *BaseRepository[T, U]) UpdateMany(ctx context.Context, entities []U) error {
	if len(entities) == 0 {
		return nil
	}

	ops := make([]mongo.WriteModel, 0, len(entities))
	for _, ent := range entities {
		r.fillTimeStamp(ctx, ent, false)
		op := mongo.NewReplaceOneModel()
		op.SetFilter(bson.M{"_id": ent.RawID()})
		op.SetReplacement(ent)
		ops = append(ops, op)
	}

	res, err := r.Collection().BulkWrite(ctx, ops)
	if err == nil && res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return err
}

func (r *BaseRepository[T, U]) UpdateOne(ctx context.Context, filter interface{}, update interface{}, options ...*repository.QueryOptions) error {
	result, err := r.Collection().UpdateOne(ctx, r.where(filter, options...), update)
	if err != nil {
		return wrapErr(err)
	}
	if result.MatchedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *BaseRepository[T, U]) UpdateAndGetByID(ctx context.Context, updateEntity U, returnNewDocument bool, queryOpt ...*repository.QueryOptions) (U, error) {
	opt := options.FindOneAndUpdate()
	if returnNewDocument {
		opt.SetReturnDocument(options.After)
	} else {
		opt.SetReturnDocument(options.Before)
	}
	r.fillTimeStamp(ctx, updateEntity, false)

	res := r.Collection().FindOneAndUpdate(ctx, r.where(bson.M{"_id": updateEntity.RawID()}, queryOpt...), bson.M{"$set": updateEntity}, opt)
	if err := wrapErr(res.Err()); err != nil {
		return nil, err
	}
	var result T
	if err := res.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *BaseRepository[T, U]) Upsert(ctx context.Context, entity U, opt ...*repository.QueryOptions) error {
	return r.UpsertOne(ctx, r.where(bson.M{"_id": entity.RawID()}, opt...), entity)
}

func (r *BaseRepository[T, U]) UpsertMany(ctx context.Context, entities []U, opt ...*repository.QueryOptions) error {
	if len(entities) == 0 {
		return nil
	}
	docs := make([]mongo.WriteModel, 0, len(entities))
	for _, entity := range entities {
		r.fillTimeStamp(ctx, entity, false)
		docs = append(docs, mongo.NewUpdateOneModel().SetFilter(bson.M{"_id": entity.RawID()}).SetUpdate(bson.M{"$set": entity}).SetUpsert(true))
	}
	_, err := r.Collection().BulkWrite(ctx, docs)
	return wrapErr(err)
}

func (r *BaseRepository[T, U]) UpsertOne(ctx context.Context, filter interface{}, update U) error {
	opts := options.UpdateOne().SetUpsert(true)
	result, err := r.Collection().UpdateOne(ctx, filter, bson.M{"$set": update}, opts)
	if err != nil {
		return wrapErr(err)
	}
	if result.MatchedCount+result.UpsertedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}

// Delete removes object by id
func (r *BaseRepository[T, U]) Delete(ctx context.Context, id interface{}, options ...*repository.QueryOptions) error {
	id = repository.ToRawID[T, U](id)

	if options != nil {
		if !options[0].Archived {
			_, err := r.Collection().DeleteOne(ctx, bson.M{"_id": id})
			return wrapErr(err)
		}
	}

	return r.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{BsonFieldNameDeleteTime: time.Now(), BsonFieldNameArchived: true}})
}

// DeleteMany removes objects
func (r *BaseRepository[T, U]) DeleteMany(ctx context.Context, filter interface{}, options ...*repository.QueryOptions) error {
	if options != nil {
		if !options[0].Archived {
			_, err := r.Collection().DeleteMany(ctx, filter)
			return wrapErr(err)
		}
	}
	_, err := r.Collection().UpdateMany(ctx, filter, bson.M{"$set": bson.M{BsonFieldNameDeleteTime: time.Now(), BsonFieldNameArchived: true}})
	return wrapErr(err)
}

func (r *BaseRepository[T, U]) fillTimeStamp(ctx context.Context, e repository.IEntityID, fillCreateTime bool) {
	if entityTimestamp, ok := e.(IEntityTimeStamp); ok {
		now := time.Now().UTC()
		var userID any
		if user, ok := ctx.Value(UserEntryCtxKey).(repository.IEntityID); ok {
			userID = user.RawID()
		}

		entityTimestamp.SetUpdateFields(now, userID)
		if fillCreateTime {
			entityTimestamp.SetCreateFields(now, userID)
		} else {
			projection := bson.D{
				{Key: BsonFieldNameCreateUserID, Value: 1},
				{Key: BsonFieldNameCreateTime, Value: 1},
			}

			result := r.Collection().FindOne(ctx,
				bson.M{"_id": e.RawID()},
				options.FindOne().SetProjection(projection))

			var entity EntityTimeStamp
			if err := result.Decode(&entity); err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				log.FromContext(ctx).InfoWithFields(log.Fields{
					"err": err,
					"id":  e.RawID(),
				}, "error decoding entity timestamp")
			}
			if entity.CreateTime == nil {
				entity.CreateTime = &now
				entity.CreateUserID = userID
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

func wrapErr(err error) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return repository.ErrNotFound
	}
	if mongo.IsDuplicateKeyError(err) {
		return fmt.Errorf("%w: %w", repository.ErrAlreadyExists, err)
	}
	return err
}
