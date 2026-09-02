package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/storage"
	"github.com/golang-migrate/migrate/v4"
	mm "github.com/golang-migrate/migrate/v4/database/clickhouse"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/stub"
)

type Auth = clickhouse.Auth

type Storage struct {
	cfg Config

	Db      *sql.DB
	native  driver.Conn
	isReady atomic.Int32
	doOnce  sync.Once

	migrations  *source.Migrations
	clusterName string
}

func NewStorage(config Config) *Storage {
	return &Storage{cfg: config, migrations: source.NewMigrations()}
}

func (s *Storage) Auth() Auth {
	return s.cfg.options.Auth
}

func (s *Storage) InCluster(clusterName string) *Storage {
	s.clusterName = clusterName
	return s
}

func (s *Storage) ClusterName() string {
	return s.clusterName
}

func (s *Storage) IsReplicationEnabled() bool {
	if len(s.clusterName) > 0 {
		return true
	}
	_, err := s.Db.Exec("SELECT * FROM system.zookeeper WHERE path = '/' LIMIT 0")
	return err == nil
}

// AddMigration adds a migration. All migrations should be added before db.Connect
func (db *Storage) AddMigration(version uint, migrationUp, migrationDown string) {
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Up, Identifier: migrationUp})
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Down, Identifier: migrationDown})
}

func (db *Storage) Connect() {
	connect := clickhouse.OpenDB(db.cfg.options)
	if err := connect.Ping(); err != nil {
		db.LogError(err)
		return
	}

	native, err := clickhouse.Open(db.cfg.options)
	if err != nil {
		db.LogError(err)
		return
	}

	dbname := db.cfg.options.Auth.Database

	use_db := "USE " + dbname + ";"
	if _, err := connect.Exec(use_db); err != nil {
		db.LogError(err)
	}

	db.Db = connect
	db.native = native

	db.isReady.Store(1)

	log.InfoWithFields(log.Fields{"dsn": &db.cfg.Dsn}, "Clickhouse connected!")
}

func (db *Storage) RunMigrations() {
	db.doOnce.Do(func() {
		dbname := db.cfg.options.Auth.Database
		if err := db.runMigrations(dbname); err != nil {
			log.Errorf("ClickHouse migrations: %s", err)
		}
	})
}

func (db *Storage) Disconnect(ctx context.Context) error {
	err := db.Db.Close()
	if err != nil {
		return fmt.Errorf("Clickhouse disconnect failed: %w", err)
	}
	log.Info("Clickhouse disconnected.")
	return nil
}

func (db *Storage) IsReady() bool {
	return db.isReady.Load() == 1
}

func (db *Storage) LogError(err error) {
	if exception, ok := err.(*clickhouse.Exception); ok {
		log.ErrorWithFields(log.Fields{
			"exception.Code":       exception.Code,
			"exception.Message":    exception.Message,
			"exception.StackTrace": exception.StackTrace,
		}, "Failed to execute clickhouse request:")
	} else {
		log.ErrorWithFields(log.Fields{"error": err}, "Failed to execute clickhouse request:")
	}
}

func (db *Storage) runMigrations(dbname string) error {
	log.Debug("Checking clickhouse DB schema...")
	cfg := &mm.Config{DatabaseName: dbname, MultiStatementEnabled: true}

	driver, err := mm.WithInstance(db.Db, cfg)
	if err != nil {
		return fmt.Errorf("driver failed: %w", err)
	}

	src, _ := source.Open("stub://")
	src.(*stub.Stub).Migrations = db.migrations
	m, err := migrate.NewWithInstance("", src, "", driver)
	m.Log = storage.MigrateLogger{}
	if err != nil {
		return fmt.Errorf("instance failed: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("process failed: %w", err)
	}

	return nil
}

func (db *Storage) Stats(ctx context.Context) (map[string]any, error) {
	metrics := make(map[string]any)
	var errs error
	for _, query := range []string{
		"SELECT event, value FROM system.events",
		"SELECT metric, value FROM system.asynchronous_metrics",
		"SELECT metric, value FROM system.metrics",
	} {
		rows, err := db.Db.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var metricName string
			var metricValue float64
			if err = rows.Scan(&metricName, &metricValue); err != nil {
				return nil, err
			}
			metrics[metricName] = metricValue
		}
		errs = errors.Join(errs, rows.Err())
	}

	return metrics, errs
}

func (db *Storage) PrepareBatch(q string) (driver.Batch, error) {
	return db.native.PrepareBatch(context.Background(), q)
}

func (db *Storage) PrepareStaticBatch(q string) (*Batch, error) {
	return NewBatch(db.native, q)
}
