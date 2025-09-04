package mongo_cache

import (
	"context"
	"errors"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/log"
	gost_mongo "github.com/bldsoft/gost/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CacheCollectionName      = "cache"
	NamespaceExistsErrorCode = 48
)

var (
	DefaultTTL = 1 * time.Minute
)

type MongoCache struct {
	collection *mongo.Collection
	ctx        context.Context
}

type CacheItem struct {
	ID       string    `bson:"_id"`
	Value    []byte    `bson:"value"`
	ExpireAt time.Time `bson:"expireAt"`
}

func NewMongoCache(s *gost_mongo.Storage) *MongoCache {
	ctx := context.Background()
	if err := s.Db.CreateCollection(ctx, CacheCollectionName); err != nil && !hasErrorCode(err, NamespaceExistsErrorCode) {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to create mongo cache collection")
	}
	collection := s.Db.Collection(CacheCollectionName)

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expireAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}

	if _, err := collection.Indexes().CreateOne(ctx, indexModel); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "Failed to create mongo cache index")
	}

	return &MongoCache{
		collection: collection,
		ctx:        ctx,
	}
}

func (mc *MongoCache) Set(key string, value []byte) error {
	return mc.SetFor(key, value, DefaultTTL)
}

func (mc *MongoCache) SetFor(key string, value []byte, ttl time.Duration) error {
	filter := bson.D{{Key: "_id", Value: key}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "value", Value: value},
			{Key: "expireAt", Value: time.Now().Add(ttl)},
		}},
	}

	opts := options.Update().SetUpsert(true)
	_, err := mc.collection.UpdateOne(mc.ctx, filter, update, opts)
	return err
}

func (mc *MongoCache) Get(key string) ([]byte, error) {
	filter := bson.M{"$and": []bson.M{{"_id": key}, {"expireAt": bson.M{"$gt": time.Now()}}}}

	var result CacheItem
	err := mc.collection.FindOne(mc.ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, cache.ErrCacheMiss
		}
		return nil, err
	}

	return result.Value, nil
}

func (mc *MongoCache) Delete(key string) error {
	_, err := mc.collection.DeleteOne(mc.ctx, bson.M{"_id": key})
	return err
}

func (mc *MongoCache) Reset() {
	_, _ = mc.collection.DeleteMany(mc.ctx, bson.M{})
}

func hasErrorCode(err error, code int32) bool {
	var cmdErr mongo.CommandError
	if errors.As(err, &cmdErr) {
		return cmdErr.Code == code
	}
	return false
}
