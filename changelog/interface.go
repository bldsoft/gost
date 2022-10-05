package changelog

import (
	"context"

	"github.com/bldsoft/gost/repository"
)

type Filter struct {
	EntityID           string
	Collections        []string
	StartTime, EndTime int64
	UserID             string
	Operations         []Operation
}

type IChangeLogRepository interface {
	GetRecords(ctx context.Context, filter *Filter) ([]*Record, error)
	FindByID(ctx context.Context, id string, options ...*repository.QueryOptions) (*Record, error)
	FindByIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) (res []*Record, err error)
}

type IChangeLogService interface {
	GetRecords(ctx context.Context, filter *Filter) ([]*Record, error)
	FindByID(ctx context.Context, id string) (*Record, error)
	FindByIDs(ctx context.Context, ids []string, preserveOrder bool) (res []*Record, err error)
}
