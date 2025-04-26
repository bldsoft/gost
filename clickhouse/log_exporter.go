package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bldsoft/gost/entity/stat"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/utils/exporter"
	"golang.org/x/sync/errgroup"
)

var ErrLogDbNotReady = errors.New("log record db isn't ready")

const (
	LevelColumName           = "level"
	MsgColumnName            = "msg"
	ServiceColumnName        = "service"
	ServiceVersionColumnName = "service_version"
	InstanseColumnName       = "instanse"
	TimestampColumnName      = "timestamp"
	ReqIDColumnName          = "req_id"
	FieldsColumnName         = "fields"

	labelColumn  = "label"
	timeColumn   = "time"
	valuesColumn = "values"
	timesColumn  = "times"
	valueColumn  = "value"
)

type LogExporterConfig struct {
	FlushTimeMs  int64 `mapstructure:"CLICKHOUSE_LOG_EXPORT_FLUSH_TIME_MS"  description:"Max time between log exporting"`
	MaxBatchSize int64 `mapstructure:"CLICKHOUSE_LOG_EXPORT_MAX_BATCH_SIZE" description:"Max batch size for log insert query"`
	ChanBufSize  int   `mapstructure:"CLICKHOUSE_LOG_EXPORT_CHAN_BUF_SIZE"  description:"-"`

	TableName string `mapstructure:"CLICKHOUSE_LOG_EXPORT_TABLE" description:"Table name for log exporting"`
}

// SetDefaults ...
func (c *LogExporterConfig) SetDefaults() {
	c.FlushTimeMs = 1000
	c.MaxBatchSize = 1000
	c.ChanBufSize = 16384
	c.TableName = "LOG_RECORDS"
}

// Validate ...
func (c *LogExporterConfig) Validate() error {
	if len(c.TableName) == 0 {
		return errors.New("log export: empty table name")
	}
	if c.MaxBatchSize <= 0 {
		return errors.New("log export: batch size isn't set")
	}
	return nil
}

type ClickHouseLogExporter struct {
	exporter.Exporter[*log.LogRecord]
	server.AsyncRunner

	BaseRepository
	config LogExporterConfig
}

func NewLogExporter(storage *Storage, cfg LogExporterConfig) *ClickHouseLogExporter {
	logExporter := &ClickHouseLogExporter{
		config:         cfg,
		BaseRepository: NewBaseRepository(storage),
	}

	go func() {
		for !storage.IsReady() {
		}

		if err := logExporter.createTableIfNotExitst(); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err}, "failed to create log table")
		}

		bufExporter := NewExporter[*chLogRecord](storage, ExporterConfig{
			cfg.TableName,
			exporter.BufferedExporterConfig{
				MaxFlushInterval: time.Duration(cfg.FlushTimeMs) * time.Millisecond,
				MaxBatchSize:     int(cfg.MaxBatchSize),
				ChanBufSize:      cfg.ChanBufSize,
				PreserveOld:      false,
				Logger:           log.Logger.WithFields(log.Fields{"exporter": "LOG_RECORDS"}),
			},
			storage.IsReadyRaw(),
		})

		logExporter.Exporter = exporter.Transform(bufExporter, formLogRecord)
		logExporter.AsyncRunner = bufExporter
	}()

	return logExporter
}

func (e *ClickHouseLogExporter) filter(filter *log.Filter) (where sq.And) {
	if filter == nil {
		return
	}
	if !filter.From.IsZero() {
		where = append(where, sq.GtOrEq{TimestampColumnName: filter.From})
	}
	if !filter.To.IsZero() {
		where = append(where, sq.LtOrEq{TimestampColumnName: filter.To})
	}
	if len(filter.Levels) > 0 {
		int8Levels := make([]int8, 0, len(filter.Levels))
		for _, lvl := range filter.Levels {
			int8Levels = append(int8Levels, int8(lvl))
		}
		where = append(where, sq.Eq{LevelColumName: int8Levels})
	}

	switch len(filter.RequestIDs) {
	case 0:
	case 1:
		where = append(where, sq.Like{ReqIDColumnName: fmt.Sprintf("%%%s%%", filter.RequestIDs[0])})
	default:
		match := fmt.Sprintf("match(%s, (?))", ReqIDColumnName)
		for i := range filter.RequestIDs {
			filter.RequestIDs[i] = regexp.QuoteMeta(filter.RequestIDs[i])
		}
		where = append(where, sq.Expr(match, strings.Join(filter.RequestIDs, "|")))
	}

	if len(filter.Instances) > 0 {
		where = append(where, sq.Eq{InstanseColumnName: filter.Instances})
	}
	if len(filter.Services) > 0 {
		where = append(where, sq.Eq{ServiceColumnName: filter.Services})
	}
	if len(filter.ServiceVersions) > 0 {
		where = append(where, sq.Eq{ServiceVersionColumnName: filter.ServiceVersions})
	}

	if filter.Search != nil && len(*filter.Search) > 0 {
		var searchFilter sq.Sqlizer
		searchFilter = e.parseExpr(*filter.Search, func(search string) sq.Sqlizer {
			search = strings.TrimSpace(search)
			return sq.Or{
				sq.Expr(
					fmt.Sprintf(`positionCaseInsensitive(%s, ?) <> 0`, MsgColumnName),
					search,
				),
				sq.Expr(
					fmt.Sprintf(`positionCaseInsensitive(%s, ?) <> 0`, FieldsColumnName),
					search,
				),
			}
		})

		if filter.TrackRequest {
			var sqWhere sq.And
			sqWhere = append(sqWhere, append([]sq.Sqlizer{searchFilter}, where...)...)

			subQuery := sq.
				Select(ReqIDColumnName).
				From(e.config.TableName).
				Where(sqWhere)

			searchFilter = sq.Expr(fmt.Sprintf("%s IN (?)", ReqIDColumnName), subQuery)
		}

		where = append(where, searchFilter)
	}

	return where
}

func (e *ClickHouseLogExporter) parseExpr(search string, makeRawExpr func(string) sq.Sqlizer) sq.Sqlizer {
	return parseHelper[sq.Or](search, "|", func(s string) sq.Sqlizer {
		return parseHelper[sq.And](s, "&", makeRawExpr)
	})
}

func parseHelper[T sq.Or | sq.And](search, op string, makeRawExpr func(string) sq.Sqlizer) sq.Sqlizer {
	exprs := split(search, op)
	exprs = slices.DeleteFunc(exprs, func(s string) bool { return len(s) == 0 })

	if len(exprs) == 1 {
		return makeRawExpr(exprs[0])
	}

	var res T
	for _, expr := range exprs {
		res = append(res, makeRawExpr(expr))
	}
	return sq.Sqlizer(res)
}

func split(search, op string) []string {
	if op == "" {
		return []string{search}
	}

	var res []string
	appendStr := func(s string) {
		res = append(res, strings.ReplaceAll(s, "\\"+op, op))
	}
	startIdx := 0
	for {
		i := strings.Index(search[startIdx:], op)
		if i < 0 {
			break
		}
		i += startIdx
		if i > 0 && search[i-1] == '\\' {
			startIdx = i + 1
			continue
		}
		appendStr(search[:i])
		search = search[i+len(op):]
		startIdx = 0
	}
	appendStr(search)
	return res
}

func (e *ClickHouseLogExporter) sort(sort log.Sort) string {
	var field string
	switch sort.Field {
	case log.SortFieldTimestamp:
		field = TimestampColumnName
	case log.SortFieldReqID:
		field = ReqIDColumnName
	case log.SortFieldLevel:
		field = LevelColumName
	default:
		field = TimestampColumnName
	}
	return fmt.Sprintf("%s %s", field, sort.Order.String())
}

func (e *ClickHouseLogExporter) countLogs(
	ctx context.Context,
	params log.LogsParams,
) (int64, error) {
	query := sq.Select("count(*)").
		From(e.config.TableName).
		Where(e.filter(params.Filter))

	row := query.RunWith(e.Storage().Db).QueryRowContext(ctx)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (e *ClickHouseLogExporter) Logs(
	ctx context.Context,
	params log.LogsParams,
) (*log.Logs, error) {
	query := sq.Select().
		Column(ServiceColumnName).
		Column(ServiceVersionColumnName).
		Column(InstanseColumnName).
		Column(fmt.Sprintf("toUnixTimestamp64Milli(%s)", TimestampColumnName)).
		Column(fmt.Sprintf("CAST(%s, 'Int8') %s", LevelColumName, LevelColumName)).
		Column(ReqIDColumnName).
		Column(MsgColumnName).
		Column(FieldsColumnName).
		From(e.config.TableName).
		Where(e.filter(params.Filter)).
		OrderBy(e.sort(params.Sort)).
		Offset(uint64(params.Offset)).
		Limit(uint64(params.Limit))

	rows, err := e.RunSelect(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs log.Logs
	for rows.Next() {
		var r log.LogRecord
		if err := rows.Scan(&r.Service, &r.ServiceVersion, &r.Instance, &r.Timestamp, &r.Level, &r.ReqID, &r.Msg, &r.Fields); err != nil {
			return nil, err
		}
		logs.Records = append(logs.Records, r)
	}

	logs.TotalCount, err = e.countLogs(ctx, params)

	return &logs, err
}

func (e *ClickHouseLogExporter) logsMetricsQuery(params *log.LogsMetricsParams) sq.SelectBuilder {
	subQuery := sq.Select().
		Column(LevelColumName+" "+labelColumn).
		Column("toStartOfInterval("+TimestampColumnName+", INTERVAL (?) second) "+timeColumn, params.StepSec).
		Column("toFloat64(COUNT(*)) "+"value").
		From(e.config.TableName).
		Where(e.filter(params.Filter)).
		GroupBy(labelColumn, timeColumn).
		OrderBy(timeColumn)

	query := sq.Select().
		Column(labelColumn).
		Column("groupArray("+timeColumn+") "+timesColumn).
		Column("groupArray("+valueColumn+") "+valuesColumn).
		Column("min("+valueColumn+") min").
		Column("max("+valueColumn+") max").
		Column("avg("+valueColumn+") avg").
		Column("round(sum("+valueColumn+")) sum").
		FromSelect(subQuery, "interval_data").
		GroupBy(labelColumn)

	return query
}

func (e *ClickHouseLogExporter) LogsMetrics(ctx context.Context, params log.LogsMetricsParams) (*stat.SeriesData, error) {
	if params.To.IsZero() {
		params.To = time.Now()
	}
	return e.getCustomChartValues(ctx, e.logsMetricsQuery(&params), params.From, params.To, time.Duration(params.StepSec)*time.Second)
}

func (e *ClickHouseLogExporter) Instances(
	ctx context.Context,
	filter log.Filter,
) ([]string, error) {
	return e.distinctValues(ctx, InstanseColumnName, filter)
}

func (e *ClickHouseLogExporter) Services(ctx context.Context, filter log.Filter) ([]string, error) {
	return e.distinctValues(ctx, ServiceColumnName, filter)
}

func (e *ClickHouseLogExporter) ServiceVersions(
	ctx context.Context,
	filter log.Filter,
) ([]string, error) {
	return e.distinctValues(ctx, ServiceVersionColumnName, filter)
}

func (e *ClickHouseLogExporter) distinctValues(
	ctx context.Context,
	column string,
	filter log.Filter,
) ([]string, error) {
	query := sq.Select("distinct " + column).
		From(e.config.TableName).
		Where(e.filter(&filter))
	rows, err := e.RunSelect(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []string
	for rows.Next() {
		var instance string
		if err := rows.Scan(&instance); err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

func (e *ClickHouseLogExporter) RequestIDs(
	ctx context.Context,
	filter log.Filter,
	limit *int,
) ([]string, int64, error) {
	g := new(errgroup.Group)
	var requestIDs []string
	var count int64
	g.Go(func() error {
		query := sq.Select("distinct " + ReqIDColumnName).
			From(e.config.TableName).
			Where(e.filter(&filter)).
			Where(sq.NotEq{ReqIDColumnName: ""})

		if limit != nil {
			query = query.Limit(uint64(*limit))
		}
		rows, err := e.RunSelect(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var requestID string
			if err := rows.Scan(&requestID); err != nil {
				return err
			}
			requestIDs = append(requestIDs, requestID)
		}
		return nil
	})
	g.Go(func() error {
		query := sq.Select("uniq(" + ReqIDColumnName + ")").
			From(e.config.TableName).
			Where(e.filter(&filter)).
			Where(sq.NotEq{ReqIDColumnName: ""})

		row := query.RunWith(e.Storage().Db).QueryRowContext(ctx)
		if err := row.Scan(&count); err != nil {
			return err
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, 0, err
	}
	return requestIDs, count, nil
}

func (e *ClickHouseLogExporter) ChangeTTL(hours int64) error {
	if !e.Storage().IsReady() {
		return ErrLogDbNotReady
	}
	_, err := e.Storage().Db.Exec(
		fmt.Sprintf(
			"ALTER TABLE %s MODIFY TTL %s + INTERVAL %d HOUR",
			e.config.TableName,
			TimestampColumnName,
			hours,
		),
	)
	return err
}

func (e *ClickHouseLogExporter) createTableIfNotExitst() error {
	engine := "MergeTree"
	if e.Storage().IsReplicationEnabled() {
		engine = "ReplicatedMergeTree"
	}

	_, err := e.Storage().Db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			`+ServiceColumnName+` LowCardinality(String),
			`+ServiceVersionColumnName+` LowCardinality(String),
			`+InstanseColumnName+` LowCardinality(String),
			`+TimestampColumnName+` DateTime64(6),
			`+LevelColumName+` Enum8('DEBUG'=0, 'INFO'=1, 'WARN'=2, 'ERROR'=3, 'FATAL'=4, 'PANIC'=5, 'TRACE'=-1),
			`+ReqIDColumnName+` String,
			`+MsgColumnName+` String,
			`+FieldsColumnName+` String
	) 
	ENGINE = %s
	PARTITION BY toMonday(`+TimestampColumnName+`)
	TTL `+`toDateTime(`+TimestampColumnName+`) + INTERVAL 1 WEEK 
	ORDER BY (`+strings.Join([]string{
		"CAST(" + LevelColumName + ",'Int8')",
		ServiceColumnName,
		InstanseColumnName,
		// ServiceVersionColumnName,
		"toDateTime(" + TimestampColumnName + ")",
	}, ",")+`)`, e.config.TableName, engine))
	return err
}

func (e *ClickHouseLogExporter) IsReady() bool {
	return e.db.IsReady()
}

type chLogRecord struct {
	Service        string    `json:"service,omitempty"        ch:"service"`
	ServiceVersion string    `json:"serviceVersion,omitempty" ch:"service_version"`
	Instance       string    `json:"instance,omitempty"       ch:"instanse"`
	Timestamp      time.Time `json:"timestamp,omitempty"      ch:"timestamp"`
	Level          int8      `json:"level,string,omitempty"   ch:"level"`
	ReqID          string    `json:"reqID,omitempty"          ch:"req_id"`
	Msg            string    `json:"msg,omitempty"            ch:"msg"`
	Fields         []byte    `json:"fields,omitempty"         ch:"fields"` // json
}

func formLogRecord(r *log.LogRecord) *chLogRecord {
	return &chLogRecord{
		Service:        r.Service,
		ServiceVersion: r.ServiceVersion,
		Instance:       r.Instance,
		Timestamp:      time.UnixMicro(r.Timestamp).UTC(),
		Level:          int8(r.Level),
		ReqID:          r.ReqID,
		Msg:            r.Msg,
		Fields:         r.Fields,
	}
}
