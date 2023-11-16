package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bldsoft/gost/log"
	"golang.org/x/sync/errgroup"
)

var (
	ErrLogDbNotReady = errors.New("log record db isn't ready")
)

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
	TableName    string `mapstructure:"CLICKHOUSE_LOG_EXPORT_TABLE" description:"Table name for log exporting"`
	FlushTimeMs  int64  `mapstructure:"CLICKHOUSE_LOG_EXPORT_FLUSH_TIME_MS" description:"Max time between log exporting"`
	MaxBatchSize int64  `mapstructure:"CLICKHOUSE_LOG_EXPORT_MAX_BATCH_SIZE" description:"Max batch size for log insert query"`
	ChanBufSize  int64  `mapstructure:"CLICKHOUSE_LOG_EXPORT_CHAN_BUF_SIZE" description:"-"`
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
	storage *Storage
	recordC chan *log.LogRecord

	stop    chan struct{}
	stopped chan struct{}
	config  LogExporterConfig

	records []*log.LogRecord
}

func NewLogExporter(storage *Storage, cfg LogExporterConfig) *ClickHouseLogExporter {
	return &ClickHouseLogExporter{storage: storage, config: cfg, recordC: make(chan *log.LogRecord, cfg.ChanBufSize), records: make([]*log.LogRecord, 0, cfg.MaxBatchSize)}
}

func (e *ClickHouseLogExporter) WriteLogRecord(rec *log.LogRecord) error {
	select {
	case <-e.stop:
		// do nothing
	default:
		e.writeToChan(rec)
	}
	return nil
}

func (e *ClickHouseLogExporter) writeToChan(rec *log.LogRecord) {
	select {
	case e.recordC <- rec:
	default:
		// channel is full
	}
}

func (e *ClickHouseLogExporter) Run() error {
	e.stop = make(chan struct{})
	e.stopped = make(chan struct{})
	defer close(e.stopped)

	if err := e.createTableIfNotExitst(); err != nil {
		log.Logger.ErrorWithFields(log.Fields{"err": err}, "failed to create log table")
	}

	flushInterval := time.Duration(e.config.FlushTimeMs) * time.Millisecond
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	last_flush := time.Now()

	tryFlush := func() bool {
		if len(e.records) == 0 {
			return true
		}

		if err := e.insertMany(e.records); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err, "current batch": len(e.records), "queued": len(e.recordC)}, "failed to export log records")
			return false
		}
		// log.Logger.TraceWithFields(log.Fields{"record count": len(e.records)}, "log exported")
		e.records = e.records[:0]
		last_flush = time.Now()
		return true
	}

	mustFlush := func() {
		for !tryFlush() {
			<-ticker.C
		}
	}

	for {
		select {
		case record := <-e.recordC:
			e.records = append(e.records, record)
			if len(e.records) == cap(e.records) {
				mustFlush()
			}
		case <-ticker.C:
			if time.Since(last_flush) >= flushInterval {
				mustFlush()
			}
		case <-e.stop:
			close(e.recordC)
			for record := range e.recordC {
				e.records = append(e.records, record)
			}
			tryFlush()
			return nil
		}
	}
}

func (s *ClickHouseLogExporter) Stop(ctx context.Context) error {
	close(s.stop)
	select {
	case <-s.stopped:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (e *ClickHouseLogExporter) insertMany(records []*log.LogRecord) error {
	if !e.storage.IsReady() {
		return ErrLogDbNotReady
	}

	tx, err := e.storage.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (%s,%s,%s,%s,%s,%s,%s,%s) VALUES (?,?,?,?,?,?,?,?)",
		e.config.TableName, ServiceColumnName, ServiceVersionColumnName, InstanseColumnName, TimestampColumnName, LevelColumName, ReqIDColumnName, MsgColumnName, FieldsColumnName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		if _, err := stmt.Exec(r.Service, r.ServiceVersion, r.Instance, time.UnixMicro(r.Timestamp), int8(r.Level), r.ReqID, r.Msg, r.Fields); err != nil {
			return err
		}
	}

	return tx.Commit()
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
			sq.NotEq{fmt.Sprintf(`positionCaseInsensitive(%s, '%s')`, MsgColumnName, *filter.Search): 0},
			sq.NotEq{fmt.Sprintf(`positionCaseInsensitive(%s, '%s')`, FieldsColumnName, *filter.Search): 0},
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

func (e *ClickHouseLogExporter) countLogs(ctx context.Context, params log.LogsParams) (int64, error) {
	query := sq.Select("count(*)").
		From(e.config.TableName).
		Where(e.filter(params.Filter))

	row := query.RunWith(e.storage.Db).QueryRowContext(ctx)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (e *ClickHouseLogExporter) Logs(ctx context.Context, params log.LogsParams) (*log.Logs, error) {
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

	rows, err := query.RunWith(e.storage.Db).Query()
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

func (e *ClickHouseLogExporter) Instances(ctx context.Context, filter log.Filter) ([]string, error) {
	return e.distinctValues(ctx, InstanseColumnName, filter)
}

func (e *ClickHouseLogExporter) Services(ctx context.Context, filter log.Filter) ([]string, error) {
	return e.distinctValues(ctx, ServiceColumnName, filter)
}

func (e *ClickHouseLogExporter) ServiceVersions(ctx context.Context, filter log.Filter) ([]string, error) {
	return e.distinctValues(ctx, ServiceVersionColumnName, filter)
}

func (e *ClickHouseLogExporter) distinctValues(ctx context.Context, column string, filter log.Filter) ([]string, error) {
	query := sq.Select("distinct " + column).
		From(e.config.TableName).
		Where(e.filter(&filter))
	rows, err := query.RunWith(e.storage.Db).Query()
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

func (e *ClickHouseLogExporter) RequestIDs(ctx context.Context, filter log.Filter, limit *int) ([]string, int64, error) {
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
		rows, err := query.RunWith(e.storage.Db).Query()
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

		row := query.RunWith(e.storage.Db).QueryRowContext(ctx)
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
	_, err := e.storage.Db.Exec(fmt.Sprintf("ALTER TABLE %s MODIFY TTL %s + INTERVAL %d HOUR", e.config.TableName, TimestampColumnName, hours))
	return err
}

func (e *ClickHouseLogExporter) createTableIfNotExitst() error {
	engine := "MergeTree"
	if e.storage.IsReplicationEnabled() {
		engine = "ReplicatedMergeTree"
	}

	_, err := e.storage.Db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
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
