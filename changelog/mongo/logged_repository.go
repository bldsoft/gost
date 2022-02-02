package mongo

import (
	"context"

	"github.com/bldsoft/gost/changelog"
	"github.com/bldsoft/gost/mongo"

	// "go.mongodb.org/mongo-driver/bson"
	gost "github.com/bldsoft/gost/repository"
)

type LoggedRepository[T mongo.IEntityID] struct {
	*mongo.Repository[T]
	wrapper *changelog.RepositoryWrapper[T]
}

func NewLoggedRepository[T mongo.IEntityID](db *mongo.MongoDb, collectionName string, changeLogRep *ChangeLogRepository) *LoggedRepository[T] {
	rep := mongo.NewRepository[T](db, collectionName)
	return &LoggedRepository[T]{
		rep,
		changelog.Wrap[T](rep, changeLogRep),
	}
}

func (r *LoggedRepository[T]) Insert(ctx context.Context, item T) error {
	return r.wrapper.Insert(ctx, item)
}

func (r *LoggedRepository[T]) Update(ctx context.Context, item T) error {
	return r.wrapper.Update(ctx, item)
}

func (r *LoggedRepository[T]) Delete(ctx context.Context, item T, options ...*gost.QueryOptions) error {
	return r.wrapper.Delete(ctx, item, options...)
}
