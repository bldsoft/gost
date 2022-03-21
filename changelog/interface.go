package changelog

import (
	"context"
)

type Filter struct {
	EntityID           string
	Collections        []string
	StartTime, EndTime int64
}

type IChangeLogRepository interface {
	GetRecords(ctx context.Context, filter *Filter) ([]*Record, error)
	GetByID(ctx context.Context, id interface{}) (*Record, error)
	GetByIDs(ctx context.Context, ids []interface{}) (res []*Record, err error)
}

type IChangeLogService interface {
	GetRecords(ctx context.Context, filter *Filter) ([]*Record, error)
	GetByID(ctx context.Context, id interface{}) (*Record, error)
	GetByIDs(ctx context.Context, ids []interface{}) (res []*Record, err error)
}
