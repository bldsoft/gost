package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bldsoft/gost/log"
)

var (
	ErrLogDbNotReady = errors.New("log record db isn't ready")
)

const (
	LevelColumName      = "level"
	MsgColumnName       = "msg"
	InstanseColumnName  = "instanse"
	TimestampColumnName = "timestamp"
	ReqIDColumnName     = "req_id"
	FieldsColumnName    = "fields"
)

type LogExporterConfig struct {
	FlushTimeMs  int64 `mapstructure:"CLICKHOUSE_LOG_EXPORT_FLUSH_TIME_MS"`
	MaxBatchSize int64 `mapstructure:"CLICKHOUSE_LOG_EXPORT_MAX_BATCH_SIZE"`
	ChanBufSize  int64 `mapstructure:"CLICKHOUSE_LOG_EXPORT_CHAN_BUF_SIZE"`

	TableName        string `mapstructure:"LOG_EXPORT_CLICKHOUSE_TABLE"`
	AllowReplication bool   `mapstructure:"CLICKHOUSE_REPLICATION_ENABLED"`
}

// SetDefaults ...
func (c *LogExporterConfig) SetDefaults() {
	c.FlushTimeMs = 1000
	c.MaxBatchSize = 1000
	c.ChanBufSize = 4096
	c.TableName = "LOG_RECORDS"
	c.AllowReplication = false
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
	config LogExporterConfig

	storage *Storage

	recordC chan *log.LogRecord
	records []*log.LogRecord

	stop    chan struct{}
	stopped chan struct{}
}

func NewLogExporter(storage *Storage, cfg LogExporterConfig) *ClickHouseLogExporter {
	return &ClickHouseLogExporter{storage: storage, config: cfg, recordC: make(chan *log.LogRecord, cfg.ChanBufSize), records: make([]*log.LogRecord, 0, cfg.MaxBatchSize)}
}

func (e *ClickHouseLogExporter) WriteLogRecord(rec *log.LogRecord) error {
	select {
	case <-e.stop:
		// do nothing
	default:
		e.recordC <- rec
	}
	return nil
}

func (e *ClickHouseLogExporter) Run() error {
	e.stop = make(chan struct{})
	e.stopped = make(chan struct{})
	defer close(e.stopped)

	if err := e.createTableIfNotExitst(); err != nil {
		log.Logger.ErrorWithFields(log.Fields{"err": err}, "failed to create log table")
	}

	ticker := time.NewTicker(time.Duration(e.config.FlushTimeMs) * time.Millisecond)
	defer ticker.Stop()

	flush := func() bool {
		if len(e.records) == 0 {
			return true
		}

		if err := e.insertMany(e.records); err != nil {
			log.Logger.ErrorWithFields(log.Fields{"err": err}, "failed to export log records")
			return false
		}
		log.Logger.TraceWithFields(log.Fields{"record count": len(e.records)}, "log exported")
		e.records = e.records[:0]
		return true
	}

	for {
		select {
		case record := <-e.recordC:
			if len(e.records) < cap(e.records) || flush() {
				e.records = append(e.records, record)
			}
		case <-ticker.C:
			flush()
		case <-e.stop:
			close(e.recordC)
			for record := range e.recordC {
				e.records = append(e.records, record)
			}
			flush()
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

	stmt, err := tx.Prepare(fmt.Sprintf("INSERT INTO %s (%s,%s,%s,%s,%s,%s) VALUES (?,?,?,?,?,?)", e.config.TableName,
		InstanseColumnName, TimestampColumnName, LevelColumName, ReqIDColumnName, MsgColumnName, FieldsColumnName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, record := range records {
		if _, err := stmt.Exec(record.Instanse, record.Timestamp, record.Level, record.ReqID, record.Msg, record.Fields); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (e *ClickHouseLogExporter) createTableIfNotExitst() error {
	engine := "MergeTree"
	if e.config.AllowReplication {
		engine = "ReplicatedMergeTree"
	}

	_, err := e.storage.Db.Exec(fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			`+InstanseColumnName+` LowCardinality(String),
			`+TimestampColumnName+` DateTime,
			`+LevelColumName+` Enum8('DEBUG'=0, 'INFO'=1, 'WARN'=2, 'ERROR'=3, 'FATAL'=4, 'PANIC'=5, 'TRACE'=-1),
			`+ReqIDColumnName+` String,
			`+MsgColumnName+` String,
			`+FieldsColumnName+` String
	) 
	ENGINE = %s
	PARTITION BY toYYYYMM(`+TimestampColumnName+`)
	ORDER BY (`+strings.Join([]string{TimestampColumnName, InstanseColumnName, ReqIDColumnName}, ",")+`)`, e.config.TableName, engine))
	return err
}
