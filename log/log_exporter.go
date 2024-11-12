package log

import (
	"context"
	"time"

	"github.com/bldsoft/gost/entity/stat"
	"github.com/bldsoft/gost/repository"
)

//go:generate go run github.com/dmarkham/enumer@latest -gqlgen -type SortField --output sort_field_enum.go --trimprefix "SortField"

type SortField int

const (
	SortFieldTimestamp SortField = iota
	SortFieldLevel
	SortFieldReqID
)

type LogExporterConfig struct {
	Service  string
	Instance string
}

type LogExporter interface {
	Export(r ...*LogRecord) (n int, err error)
	Logs(ctx context.Context, params LogsParams) (*Logs, error)
	LogsMetrics(ctx context.Context, params LogsMetricsParams) (*stat.SeriesData, error)
	Instances(ctx context.Context, filter Filter) ([]string, error)
	Services(ctx context.Context, filter Filter) ([]string, error)
	ServiceVersions(ctx context.Context, filter Filter) ([]string, error)
	RequestIDs(ctx context.Context, filter Filter, limit *int) ([]string, int64, error)
}

type LogsMetricsParams struct {
	*Filter
	StepSec float64 `json:"stepSec,omitempty" schema:"stepSec,omitempty"`
}

type LogsParams struct {
	Offset  int `json:"offset,omitempty" schema:"offset,omitempty"`
	Limit   int `json:"limit,omitempty"  schema:"limit,omitempty"`
	*Filter `    json:"filter,omitempty" schema:",omitempty"`
	Sort    `    json:"sort,omitempty"   schema:",omitempty"`
}

type Filter struct {
	Services        []string  `json:"services,omitempty"        schema:"services,omitempty"`
	ServiceVersions []string  `json:"serviceVersions,omitempty" schema:"serviceVersions,omitempty"`
	Instances       []string  `json:"instances,omitempty"       schema:"instances,omitempty"`
	Search          *string   `json:"search,omitempty"          schema:"search,omitempty"`
	From            time.Time `json:"from"                      schema:"from,omitempty"`
	To              time.Time `json:"to"                        schema:"to,omitempty"`
	RequestIDs      []string  `json:"requestIDs,omitempty"      schema:"reqID,omitempty"`
	Levels          []Level   `json:"levels,omitempty"          schema:"levels,omitempty"`
	TrackRequest    bool      `json:"trackRequest,omitempty" 		schema:"trackRequest,omitempty"`
}

type Sort struct {
	Field SortField            `json:"field,omitempty" schema:"field,omitempty"`
	Order repository.SortOrder `json:"order,omitempty" schema:"order,omitempty"`
}

type Logs struct {
	Records    []LogRecord `json:"records"`
	TotalCount int64       `json:"totalCount"`
}
