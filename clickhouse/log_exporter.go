package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
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

	config  LogExporterConfig
	storage *Storage
}

func NewLogExporter(storage *Storage, cfg LogExporterConfig) *ClickHouseLogExporter {
	logExporter := &ClickHouseLogExporter{
		config:  cfg,
		storage: storage,
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
	})

	logExporter.Exporter = exporter.Transform(bufExporter, formLogRecord)
	logExporter.AsyncRunner = bufExporter

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
	if filter.Search != nil && len(*filter.Search) > 0 {
		where = append(where, sq.Or{
			sq.Expr(
				fmt.Sprintf(`positionCaseInsensitive(%s, ?) <> 0`, MsgColumnName),
				*filter.Search,
			),
			sq.Expr(
				fmt.Sprintf(`positionCaseInsensitive(%s, ?) <> 0`, FieldsColumnName),
				*filter.Search,
			),
		})
	}

	switch len(filter.RequestIDs) {
	case 0:
	case 1:
		where = append(where, sq.Like{ReqIDColumnName: fmt.Sprintf("%%%s%%", filter.RequestIDs[0])})
	default:
		match := fmt.Sprintf("match(%s, (?))", ReqIDColumnName)
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
	return where
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

	q, args, _ := query.ToSql()
	row := e.storage.Native.QueryRow(ctx, q, args...)
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

	q, args, _ := query.ToSql()
	rows, err := e.storage.Native.Query(ctx, q, args...)
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
	q, args, _ := query.ToSql()
	rows, err := e.storage.Native.Query(ctx, q, args...)
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

		q, args, _ := query.ToSql()
		rows, err := e.storage.Native.Query(ctx, q, args)
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

		q, args, _ := query.ToSql()
		row := e.storage.Native.QueryRow(ctx, q, args...)
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
	if !e.storage.IsReady() {
		return ErrLogDbNotReady
	}

	err := e.storage.Native.Exec(context.Background(), fmt.Sprintf(
		"ALTER TABLE %s MODIFY TTL %s + INTERVAL %d HOUR",
		e.config.TableName,
		TimestampColumnName,
		hours,
	))
	return err
}

func (e *ClickHouseLogExporter) createTableIfNotExitst() error {
	engine := "MergeTree"
	if e.storage.IsReplicationEnabled() {
		engine = "ReplicatedMergeTree"
	}

	err := e.storage.Native.Exec(context.Background(), fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
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
	PARTITION BY toYYYYMM(`+TimestampColumnName+`)
	TTL `+`toDateTime(`+TimestampColumnName+`) + INTERVAL 1 MONTH 
	ORDER BY (`+strings.Join([]string{
		"CAST(" + LevelColumName + ",'Int8')",
		ServiceColumnName,
		InstanseColumnName,
		// ServiceVersionColumnName,
		"toDateTime(" + TimestampColumnName + ")",
	}, ",")+`)`, e.config.TableName, engine))
	return err
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
