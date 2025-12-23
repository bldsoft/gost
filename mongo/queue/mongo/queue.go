package mongo

import (
	"context"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
)

type Item[T any] struct {
	mongo.EntityID `bson:",inline"`
	Value          T

	CreatedAt        time.Time `bson:"createdAt,omitempty"`
	ProcessStartedAt time.Time `bson:"processStartedAt,omitempty"`
	DoneAt           time.Time `bson:"doneAt,omitempty"`
}

func (it Item[T]) MarshalBSON() ([]byte, error) {
	tmp := struct {
		mongo.EntityID `bson:",inline"`
		Value          T

		CreatedAt        *time.Time `bson:"createdAt,omitempty"`
		ProcessStartedAt *time.Time `bson:"processStartedAt,omitempty"`
		DoneAt           *time.Time `bson:"doneAt,omitempty"`
	}{
		EntityID: it.EntityID,
		Value:    it.Value,
	}
	if !it.CreatedAt.IsZero() {
		tmp.CreatedAt = &it.CreatedAt
	}
	if !it.ProcessStartedAt.IsZero() {
		tmp.ProcessStartedAt = &it.ProcessStartedAt
	}
	if !it.DoneAt.IsZero() {
		tmp.DoneAt = &it.DoneAt
	}
	return bson.Marshal(tmp)
}

type Config struct {
	ProcessTimeout    time.Duration
	OrderedProcessing bool
}

var DefaultConfig = Config{
	ProcessTimeout:    1 * time.Minute,
	OrderedProcessing: true,
}

type Queue[T any] struct {
	cfg        Config
	repository mongo.Repository[Item[T], *Item[T]]
}

func NewQueue[T any](db *mongo.Storage, collectionName string, cfg Config) *Queue[T] {
	if cfg.ProcessTimeout < 0 {
		cfg.ProcessTimeout = DefaultConfig.ProcessTimeout
	}

	return &Queue[T]{
		cfg:        cfg,
		repository: mongo.NewRepository[Item[T]](db, collectionName),
	}
}

func (q *Queue[T]) Enqueue(ctx context.Context, entity ...T) error {
	now := time.Now()
	items := make([]*Item[T], 0, len(entity))
	for _, e := range entity {
		items = append(items, &Item[T]{
			Value:     e,
			CreatedAt: now,
		})
	}
	return q.repository.InsertMany(ctx, items)
}

func (q *Queue[T]) Dequeue(ctx context.Context) (_ *Item[T], _ error) {
	items, err := q.DequeueMany(ctx, 1)
	if err != nil {
		return nil, err
	}
	return items[0], nil
}

func (q *Queue[T]) DequeueMany(ctx context.Context, n int) (_ []*Item[T], _ error) {
	now := time.Now()
	filter := bson.D{
		{Key: "doneAt", Value: bson.M{"$exists": false}},
	}

	if !q.cfg.OrderedProcessing {
		filter = append(filter, bson.E{Key: "$or", Value: []bson.M{
			{"processStartedAt": bson.M{"$exists": false}},
			{"processStartedAt": bson.M{"$lt": now.Add(-q.cfg.ProcessTimeout)}},
		}})
	}

	itemsI, err := q.repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		items, err := q.repository.Find(ctx, filter, &repository.QueryOptions{
			Limit: int64(n),
			Sort:  repository.Sort().Asc("createdAt"),
		})
		if err != nil || len(items) == 0 {
			return nil, err
		}
		if q.cfg.OrderedProcessing && !items[0].ProcessStartedAt.IsZero() {
			return nil, nil
		}

		ids := make([]interface{}, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.RawID())

			item.ProcessStartedAt = now
		}
		_, err = q.repository.Collection().UpdateMany(ctx, bson.M{"_id": bson.M{"$in": ids}}, bson.M{
			"$set": bson.M{"processStartedAt": now},
		})
		if err != nil {
			return nil, err
		}
		return items, nil
	})
	if err != nil || itemsI == nil {
		return nil, err
	}
	items := itemsI.([]*Item[T])
	return items, nil

}

func (q *Queue[T]) AckItems(ctx context.Context, items ...*Item[T]) error {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.EntityID.StringID())
	}
	return q.Ack(context.Background(), ids...)
}

func (q *Queue[T]) NackItems(ctx context.Context, items ...*Item[T]) error {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.EntityID.StringID())
	}
	return q.Nack(context.Background(), ids...)
}

func (q *Queue[T]) Ack(ctx context.Context, id ...string) error {
	return q.updateItems(ctx, id, bson.M{"$set": bson.M{"doneAt": time.Now()}})
}

func (q *Queue[T]) Nack(ctx context.Context, id ...string) error {
	return q.updateItems(ctx, id, bson.M{"$unset": bson.M{"processStartedAt": ""}})
}

func (q *Queue[T]) updateItems(ctx context.Context, ids []string, update bson.M) error {
	if len(ids) == 0 {
		return nil
	}
	rawIDs := repository.StringsToRawIDs[Item[T]](ids)
	updateRes, err := q.repository.Collection().UpdateMany(ctx, bson.M{"_id": bson.M{"$in": rawIDs}}, update)
	if err != nil {
		return err
	}
	if updateRes.MatchedCount != int64(len(rawIDs)) {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"matchedCount": updateRes.MatchedCount, "total": len(rawIDs)}, "some items not found")
		return nil
	}
	return err
}
