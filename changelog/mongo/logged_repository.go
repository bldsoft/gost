package mongo

import (
	"context"
	"encoding/json"

	"github.com/bldsoft/gost/changelog"
	jsonpatch "github.com/evanphx/json-patch"

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
	rec, err := newRecord(ctx, r.Name(), changelog.Create)
	if err != nil {
		return err
	}

	_, err = r.Repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		entity.SetChangeID(rec.GetID())
		if err := r.Repository.Insert(ctx, entity); err != nil {
			return nil, err
		}
		rec.Record.EntityID = entity.GetID()
		rec.SetData(entity)
		return nil, r.changeLogRep.Insert(ctx, rec)
	})
	return err
}

func (r *LoggedRepository[T, U]) getDiff(old U, new U) ([]byte, error) {
	oldData, err := json.Marshal(old)
	if err != nil {
		return nil, err
	}
	newData, err := json.Marshal(new)
	if err != nil {
		return nil, err
	}

	patch, err := jsonpatch.CreateMergePatch(oldData, newData)
	if err != nil {
		return nil, err
	}
	return patch, nil
}

func (r *LoggedRepository[T, U]) Update(ctx context.Context, entity U) error {
	rec, err := newRecord(ctx, r.Name(), changelog.Update)
	if err != nil {
		return err
	}

	_, err = r.Repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		entity.SetChangeID(rec.GetID())
		oldEntity, err := r.Repository.UpdateAndGetByID(ctx, entity, false)
		if err != nil {
			return nil, err
		}

		rec.Record.EntityID = entity.GetID()
		oldEntity.SetChangeID(rec.GetID())
		data, err := r.getDiff(oldEntity, entity)
		if err != nil {
			return nil, err
		}
		rec.Record.Data = string(data)

		return nil, r.changeLogRep.Insert(ctx, rec)
	})
	return err
}

func (r *LoggedRepository[T, U]) Delete(ctx context.Context, id interface{}, options ...*repository.QueryOptions) error {
	rec, err := newRecord(ctx, r.Name(), changelog.Delete)
	if err != nil {
		return err
	}

	_, err = r.Repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		if err := r.Repository.Delete(ctx, id, options...); err != nil {
			return nil, err
		}

		rec.Record.EntityID = id
		if entity, err := r.Repository.FindByID(ctx, id); err == nil {
			entity.SetChangeID(rec.GetID())
			if err := r.Repository.Update(ctx, entity); err != nil {
				return nil, err
			}
			rec.Record.SetData(entity)
		}
		return nil, r.changeLogRep.Insert(ctx, rec)
	})
	return err
}
