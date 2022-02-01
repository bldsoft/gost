package changelog

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/bldsoft/gost/repository"
	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5/middleware"
)

var UserEntryCtxKey = &utils.ContextKey{"UserEntry"}
var UserNotFound = errors.New("User isn't found in context")

type RepositoryWrapper[T EntityID] struct {
	mainRep IRepository[T]
	changeLogRep IChangeLogRepository
}

func Wrap[T EntityID](mainRep IRepository[T], changeLogRep IChangeLogRepository) *RepositoryWrapper[T] {
	return &RepositoryWrapper[T]{mainRep: mainRep, changeLogRep: changeLogRep}	
}

func (r *RepositoryWrapper[T]) Name() string {
	return r.mainRep.Name()
}

func (r *RepositoryWrapper[T]) createRecord(ctx context.Context, op Operation, entity T) (*Record, error) {
	user, ok := ctx.Value(UserEntryCtxKey).(EntityID)
	if !ok {
		return nil, UserNotFound
	}

	data, err := json.Marshal(entity)
	if err != nil {
		return nil, err
	}

	return &Record{
		UserID: user.GetID(),
		Timestamp: time.Now().Unix(),
		Operation: op,
		Entity: r.mainRep.Name(),
		EntityID: entity.GetID(),
		RequestID: middleware.GetReqID(ctx),
		Data: string(data),
	}, nil
}

func (r *RepositoryWrapper[T]) Insert(ctx context.Context, entity T) error {
	if err := r.mainRep.Insert(ctx, entity); err != nil {
		return err
	}

	record, err := r.createRecord(ctx, Create, entity)
	if err != nil {
		return err
	}	
	return r.changeLogRep.Insert(ctx, record)
}

func (r *RepositoryWrapper[T]) Update(ctx context.Context, entity T) error {	
	if err := r.mainRep.Update(ctx, entity); err != nil {
		return err
	}

	record, err := r.createRecord(ctx, Update, entity)
	if err != nil {
		return err
	}	
	return r.changeLogRep.Insert(ctx, record)
}

func (r *RepositoryWrapper[T]) Delete(ctx context.Context, entity T, options ...*repository.QueryOptions) error {	
	if err := r.mainRep.Delete(ctx, entity, options...); err != nil {
		return err
	}

	record, err := r.createRecord(ctx, Delete, entity)
	if err != nil {
		return err
	}	
	return r.changeLogRep.Insert(ctx, record)
}