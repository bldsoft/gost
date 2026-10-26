package log

import (
	"context"
	"regexp"
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
	LogsPatterns(ctx context.Context, params LogsPatternsParams) (*LogPatterns, error)
	Instances(ctx context.Context, filter Filter) ([]string, error)
	Services(ctx context.Context, filter Filter) ([]string, error)
	ServiceVersions(ctx context.Context, filter Filter) ([]string, error)
	RequestIDs(ctx context.Context, filter Filter, limit *int) ([]string, int64, error)
}

type LogsMetricsParams struct {
	*Filter
	StepSec float64 `json:"stepSec,omitempty" schema:"stepSec,omitempty"`
}

type LogsPatternsParams struct {
	*Filter
	Limit   int     `json:"limit,omitempty"   schema:"limit,omitempty"`
	StepSec float64 `json:"stepSec,omitempty" schema:"stepSec,omitempty"`
}

type LogsParams struct {
	Offset  int `json:"offset,omitempty" schema:"offset,omitempty"`
	Limit   int `json:"limit,omitempty"  schema:"limit,omitempty"`
	*Filter `    json:"filter,omitempty" schema:",omitempty"`
	Sort    `    json:"sort"   schema:",omitempty"`
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
	TrackRequest    bool      `json:"trackRequest,omitempty"    schema:"trackRequest,omitempty"`
}

type Sort struct {
	Field SortField            `json:"field,omitempty" schema:"field,omitempty"`
	Order repository.SortOrder `json:"order,omitempty" schema:"order,omitempty"`
}

type Logs struct {
	Records    []LogRecord `json:"records"`
	TotalCount int64       `json:"totalCount"`
}

var Masks = []struct {
	Re   *regexp.Regexp
	Repl string
}{
	{regexp.MustCompile(`"[^"]*"`), `"…"`},
	{regexp.MustCompile(`\B/[^\s:]+`), `<path>`},
	{regexp.MustCompile(`\b0x[0-9a-fA-F]+`), `<hex>`},
	{regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`), `<uuid>`},
	{regexp.MustCompile(`[0-9]{1,3}([.][0-9]{1,3}){3}(:[0-9]+){0,1}`), `<ip>`},
	{regexp.MustCompile(`\b[0-9]+\b`), `<N>`},
	{regexp.MustCompile(`[0-9a-fA-F]{8,}`), `<hex>`},
	{regexp.MustCompile(`[0-9]+`), `<N>`},
}

func Mask(s string) string {
	for _, m := range Masks {
		s = m.Re.ReplaceAllString(s, m.Repl)
	}

	return s
}

type LogPatterns struct {
	Patterns      []LogPattern `json:"patterns"`
	TotalCount    int64        `json:"totalCount"`
	PatternsCount int64        `json:"patternsCount"`
}

type LogPattern struct {
	Level     Level     `json:"level"`
	Msg       string    `json:"msg"`
	Err       string    `json:"err,omitempty"`
	Count     int64     `json:"count"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`
	Instances []string  `json:"instances"`
	ReqIDs    []string  `json:"reqIDs"`
	Example   string    `json:"example"`
	Series    []int64   `json:"series"`
}
