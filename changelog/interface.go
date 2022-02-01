package changelog

import (
	"context"

	"github.com/bldsoft/gost/repository"
)

type EntityID interface {
	GetID() interface{}
}

type IRepository[T EntityID] interface {
	Name() string // collection name
	Insert(ctx context.Context, entity T) error
	Update(ctx context.Context, entity T, options ...*repository.QueryOptions) error
	Delete(ctx context.Context, entity T, options ...*repository.QueryOptions) error
}

type Filter struct {
	EntityID           string
	Collections        []string
	StartTime, EndTime int64
}

type IChangeLogRepository interface {
	Insert(ctx context.Context, record *Record) error
	GetRecords(ctx context.Context, filter *Filter) ([]*Record, error)
}

type IChangeLogService interface {
	GetRecords(ctx context.Context, filter *Filter) ([]*Record, error)
}
