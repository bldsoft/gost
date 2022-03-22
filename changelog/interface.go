package changelog

import (
	"context"

	"github.com/bldsoft/gost/repository"
)

type Filter struct {
	EntityID           string
	Collections        []string
	StartTime, EndTime int64
}

type IChangeLogRepository interface {
	GetRecords(ctx context.Context, filter *Filter) ([]*Record, error)
	FindByID(ctx context.Context, id interface{}, options ...*repository.QueryOptions) (*Record, error)
	FindByIDs(ctx context.Context, ids []interface{}, preserveOrder bool, options ...*repository.QueryOptions) (res []*Record, err error)
}

type IChangeLogService interface {
	GetRecords(ctx context.Context, filter *Filter) ([]*Record, error)
	FindByID(ctx context.Context, id interface{}) (*Record, error)
	FindByIDs(ctx context.Context, ids []interface{}, preserveOrder bool) (res []*Record, err error)
}
