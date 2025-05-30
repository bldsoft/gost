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
	repository.IEntityID
}
type LoggedRepository[T any, U LoggedEntity[T]] struct {
	mongo.Repository[T, U]
	changeLogRep *ChangeLogRepository
}

func NewLoggedRepository[T any, U LoggedEntity[T]](db *mongo.Storage, collectionName string, changeLogRep *ChangeLogRepository) *LoggedRepository[T, U] {
	rep := mongo.NewRepository[T, U](db, collectionName)
	_ = db.Db.CreateCollection(context.Background(), collectionName)

	return &LoggedRepository[T, U]{rep, changeLogRep}
}

func WrapRepository[T any, U LoggedEntity[T]](repo mongo.Repository[T, U], changeLogRepo *ChangeLogRepository) *LoggedRepository[T, U] {
	db := repo.Collection().Database()
	_ = db.CreateCollection(context.Background(), repo.Collection().Name())
	return &LoggedRepository[T, U]{repo, changeLogRepo}
}

func (r *LoggedRepository[T, U]) Insert(ctx context.Context, entity U) (err error) {
	rec, err := NewRecord(ctx, r.Name(), changelog.Create)
	if err != nil {
		return err
	}

	_, err = r.Repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		entity.SetChangeID(rec.StringID())
		if err := r.Repository.Insert(ctx, entity); err != nil {
			return nil, err
		}
		rec.Record.EntityID = entity.StringID()
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

func (r *LoggedRepository[T, U]) Update(ctx context.Context, entity U, opt ...*repository.QueryOptions) error {
	rec, err := NewRecord(ctx, r.Name(), changelog.Update)
	if err != nil {
		return err
	}

	_, err = r.Repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		entity.SetChangeID(rec.StringID())
		oldEntity, err := r.Repository.UpdateAndGetByID(ctx, entity, false, opt...)
		if err != nil {
			return nil, err
		}

		rec.Record.EntityID = entity.StringID()
		oldEntity.SetChangeID(rec.StringID())
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
	rec, err := NewRecord(ctx, r.Name(), changelog.Delete)
	if err != nil {
		return err
	}

	_, err = r.Repository.WithTransaction(ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		if err := r.Repository.Delete(ctx, id, options...); err != nil {
			return nil, err
		}

		rec.Record.EntityID = repository.ToStringID[T, U](id)
		if entity, err := r.Repository.FindByID(ctx, id); err == nil {
			entity.SetChangeID(rec.StringID())
			if err := r.Repository.Update(ctx, entity); err != nil {
				return nil, err
			}
			rec.Record.SetData(entity)
		}
		return nil, r.changeLogRep.Insert(ctx, rec)
	})
	return err
}
