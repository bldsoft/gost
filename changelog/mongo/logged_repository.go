package mongo

import (
	"context"

	"github.com/bldsoft/gost/changelog"

	"github.com/bldsoft/gost/mongo"
	"github.com/bldsoft/gost/repository"
	// "go.mongodb.org/mongo-driver/bson"
)

type LoggedEntity[T any] interface {
	*T
	changelog.ILoggedEntity
	mongo.IEntityID
}
type LoggedRepository[T any, U LoggedEntity[T]] struct {
	*mongo.Repository[T, U]
	changeLogRep *ChangeLogRepository
}

func NewLoggedRepository[T any, U LoggedEntity[T]](db *mongo.MongoDb, collectionName string, changeLogRep *ChangeLogRepository) *LoggedRepository[T, U] {
	rep := mongo.NewRepository[T, U](db, collectionName)
	db.AddOnConnectHandler(func() {
		// creating collection beforehand for transactions
		db.Db.CreateCollection(context.Background(), collectionName)
	})
	return &LoggedRepository[T, U]{rep, changeLogRep}
}

func (r *LoggedRepository[T, U]) Insert(ctx context.Context, entity U) (err error) {
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

func (r *LoggedRepository[T, U]) Update(ctx context.Context, entity U) error {
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

func (r *LoggedRepository[T, U]) Delete(ctx context.Context, entity U, options ...*repository.QueryOptions) error {
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
