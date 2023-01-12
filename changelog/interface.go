package changelog

import (
	"context"

	"github.com/bldsoft/gost/repository"
)

const (
	BsonFieldRecords    = "records"
	BsonFieldTotalCount = "totalCount"
)

//go:generate go run github.com/dmarkham/enumer@latest -gqlgen -type SortField --output sort_field_enum.go --trimprefix "SortField"

type SortField int

const (
	// SortFieldTimestamp is a SortField of type Timestamp.
	SortFieldTimestamp SortField = iota
	// SortFieldUser is a SortField of type User.
	SortFieldUser
	// SortFieldOperation is a SortField of type Operation.
	SortFieldOperation
	// SortFieldEntity is a SortField of type Entity.
	SortFieldEntity
)

type Filter struct {
	EntityID   string      `json:"entityID,omitempty" schema:"entityID,omitempty"`
	Entities   []string    `json:"entities,omitempty" schema:"entities,omitempty"`
	UserIDs    []string    `json:"userIDs,omitempty" schema:"userIDs,omitempty"`
	Operations []Operation `json:"actions,omitempty" schema:"actions,omitempty"`
	Search     *string     `json:"search,omitempty" schema:"search,omitempty"`
	From       *int        `json:"from,omitempty" schema:"from,omitempty"`
	To         *int        `json:"to,omitempty" schema:"to,omitempty"`
	Details    []string    `json:"details,omitempty" schema:"details,omitempty"`
}

type Sort struct {
	Field SortField            `json:"field,omitempty" schema:"field,omitempty"`
	Order repository.SortOrder `json:"order,omitempty" schema:"order,omitempty"`
}

type RecordsParams struct {
	Offset  int64 `json:"offset,omitempty" schema:"offset,omitempty"`
	Limit   int64 `json:"limit,omitempty" schema:"limit,omitempty"`
	*Filter `json:"filter,omitempty" schema:",omitempty"`
	Sort    `json:"sort,omitempty" schema:",omitempty"`
}

type Records struct {
	Records    []Record `json:"records,omitempty" bson:"records,omtempty"`
	TotalCount int64    `json:"totalCount,omitempty" bson:"totalCount,omtempty"`
}

type IChangeLogRepository interface {
	GetRecords(ctx context.Context, params *RecordsParams) (*Records, error)
	FindByID(ctx context.Context, id string, options ...*repository.QueryOptions) (*Record, error)
	FindByIDs(ctx context.Context, ids []string, preserveOrder bool, options ...*repository.QueryOptions) (res []*Record, err error)
}

type IChangeLogService interface {
	GetRecords(ctx context.Context, params *RecordsParams) (*Records, error)
	FindByID(ctx context.Context, id string) (*Record, error)
	FindByIDs(ctx context.Context, ids []string, preserveOrder bool) (res []*Record, err error)
}
