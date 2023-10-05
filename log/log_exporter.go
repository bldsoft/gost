package log

import (
	"context"
	"time"

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
	DeprecatedServiceInstance string `mapstructure:"SERVICE_NAME" description:"DEPRECATED. The name is used to identify the service in logs. " `
	Instance                  string `mapstructure:"SERVICE_INSTANCE_NAME" description:"The name is used to identify the service in logs. "`
}

// SetDefaults ...
func (c *LogExporterConfig) SetDefaults() {
	c.DeprecatedServiceInstance = "default_name"
}

// Validate ...
func (c *LogExporterConfig) Validate() error {
	var err error
	if len(c.Instance) == 0 {
		c.Instance = c.DeprecatedServiceInstance
	}
	return err
}

type LogExporter interface {
	WriteLogRecord(r *LogRecord) error
	Logs(ctx context.Context, params LogsParams) (*Logs, error)
	Instances(ctx context.Context, filter Filter) ([]string, error)
	RequestIDs(ctx context.Context, filter Filter, limit *int) ([]string, int64, error)
}

type LogsParams struct {
	Offset  int `json:"offset,omitempty" schema:"offset,omitempty"`
	Limit   int `json:"limit,omitempty" schema:"limit,omitempty"`
	*Filter `json:"filter,omitempty" schema:",omitempty"`
	Sort    `json:"sort,omitempty" schema:",omitempty"`
}

type Filter struct {
	Instances  []string  `json:"instances,omitempty" schema:"instances,omitempty"`
	Search     *string   `json:"search,omitempty" schema:"search,omitempty"`
	From       time.Time `json:"from" schema:"from,omitempty"`
	To         time.Time `json:"to" schema:"to,omitempty"`
	RequestIDs []string  `json:"requestIDs,omitempty" schema:"reqID,omitempty"`
	Levels     []Level   `json:"levels,omitempty" schema:"levels,omitempty"`
}

type Sort struct {
	Field SortField            `json:"field,omitempty" schema:"field,omitempty"`
	Order repository.SortOrder `json:"order,omitempty" schema:"order,omitempty"`
}

type Logs struct {
	Records    []LogRecord `json:"records"`
	TotalCount int64       `json:"totalCount"`
}
