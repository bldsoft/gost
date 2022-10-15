package log

import (
	"context"
	"time"
)

//go:generate go run github.com/abice/go-enum@latest -f=$GOFILE

// ENUM(Timestamp, Level, ReqID)
type SortField int

// ENUM(ASC, DESC)
type SortOrder int

type LogExporterConfig struct {
	Instanse string `mapstructure:"SERVICE_NAME" description:"The name is used to identify the service in logs"`
}

type LogExporter interface {
	WriteLogRecord(r *LogRecord) error
	Logs(ctx context.Context, params LogsParams) (*Logs, error)
}

type LogsParams struct {
	Offset  int `json:"offset,omitempty" schema:"offset,omitempty"`
	Limit   int `json:"limit,omitempty" schema:"limit,omitempty"`
	*Filter `json:"filter,omitempty" schema:",omitempty"`
	Sort    `json:"sort,omitempty" schema:",omitempty"`
}

type Filter struct {
	Search    *string   `json:"search,omitempty" schema:"search,omitempty"`
	From      time.Time `json:"from" schema:"from,omitempty"`
	To        time.Time `json:"to" schema:"to,omitempty"`
	RequestID *string   `json:"requestID,omitempty" schema:"reqID,omitempty"`
	Levels    []Level   `json:"levels,omitempty" schema:"levels,omitempty"`
}

type Sort struct {
	Field SortField `json:"field,omitempty" schema:"field,omitempty"`
	Order SortOrder `json:"order,omitempty" schema:"order,omitempty"`
}

type Logs struct {
	Records    []LogRecord `json:"records"`
	TotalCount int         `json:"totalCount"`
}
