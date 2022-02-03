package mongo

import (
	"context"

	"github.com/bldsoft/gost/changelog"

	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
	// "go.mongodb.org/mongo-driver/bson"
)

type LoggedEntity interface {
	changelog.ILoggedEntity
	mongo.IEntityID
}
type LoggedRepository[T LoggedEntity] struct {
	*mongo.Repository[T]
	changeLogRep *ChangeLogRepository
}

func NewLoggedRepository[T LoggedEntity](db *mongo.MongoDb, collectionName string, changeLogRep *ChangeLogRepository) *LoggedRepository[T] {
	rep := mongo.NewRepository[T](db, collectionName)
	return &LoggedRepository[T]{rep, changeLogRep}
}

func (r *LoggedRepository[T]) Insert(ctx context.Context, entity T) (err error) {
	rec, err := newRecord(ctx, r.Name(), changelog.Create, entity)
	if err != nil {
		return err
	}

	_, err = r.Repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		if err := r.changeLogRep.Insert(ctx, rec); err != nil {
			return nil, err
		}
		entity.SetChangeID(rec.GetID())
		return nil, r.Repository.Insert(ctx, entity)
	})
	return err
}

func (r *LoggedRepository[T]) Update(ctx context.Context, entity T) error {
	rec, err := newRecord(ctx, r.Name(), changelog.Update, entity)
	if err != nil {
		return err
	}

	_, err = r.Repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		if err := r.changeLogRep.Insert(ctx, rec); err != nil {
			return nil, err
		}
		entity.SetChangeID(rec.GetID())
		return nil, r.Repository.Update(ctx, entity)
	})
	return err
}

func (r *LoggedRepository[T]) Delete(ctx context.Context, entity T, options ...*repository.QueryOptions) error {
	rec, err := newRecord(ctx, r.Name(), changelog.Delete, entity)
	if err != nil {
		return err
	}

	_, err = r.Repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		if err := r.changeLogRep.Insert(ctx, rec); err != nil {
			return nil, err
		}
		entity.SetChangeID(rec.GetID())
		return nil, r.Repository.Delete(ctx, entity, options...)
	})
	return err
}
