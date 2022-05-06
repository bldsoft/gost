package clickhouse

import (
	"context"
	"database/sql"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/clickhouse-go"
	gost "github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/storage"
	"github.com/golang-migrate/migrate/v4"
	mm "github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/pkg/errors"

	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/stub"
)

type Clickhouse struct {
	dsn gost.ConnectionString

	Db      *sql.DB
	isReady int32
	doOnce  sync.Once

	migrations *source.Migrations
}

func NewClickhouseConnection(dsn gost.ConnectionString) *Clickhouse {
	db := Clickhouse{dsn: dsn, migrations: source.NewMigrations()}
	return &db
}

//AddMigration adds a migration. All migrations should be added before db.Connect
func (db *Clickhouse) AddMigration(version uint, migrationUp, migrationDown string) {
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Up, Identifier: migrationUp})
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Down, Identifier: migrationDown})
}

func (db *Clickhouse) Connect() {
	connect, err := sql.Open("clickhouse", db.dsn.String())

	if err != nil {
		log.ErrorWithFields(log.Fields{"dsn": &db.dsn, "error": err}, "Failed to connect clickhouse db")

		return
	}

	if err := connect.Ping(); err != nil {
		db.LogError(err)

		return
	} else {
		dbname := db.getDsnQueryParam("database")

		use_db := "USE " + dbname + ";"

		if _, err = connect.Exec(use_db); err == nil {

			db.Db = connect

			log.InfoWithFields(log.Fields{"dsn": &db.dsn}, "Clickhouse connected!")

			//run migrations only once
			db.doOnce.Do(func() {
				go func() {
					//run migrations
					if db.runMigrations(dbname) {
						atomic.StoreInt32(&db.isReady, 1)
					}
				}()
			})

			// TODO: remove it
			time.Sleep(2 * time.Second)
		} else {
			db.LogError(err)
		}

		return
	}
}

func (db *Clickhouse) Disconnect(ctx context.Context) error {
	err := db.Db.Close()
	if err != nil {
		return errors.Wrap(err, "Clickhouse disconnect failed")
	}
	log.Info("Clickhouse disconnected.")
	return nil
}

func (db *Clickhouse) IsReady() bool {
	return atomic.LoadInt32(&db.isReady) == 1
}

func (db *Clickhouse) LogError(err error) {
	if exception, ok := err.(*clickhouse.Exception); ok {
		log.ErrorWithFields(log.Fields{
			"exception.Code":       exception.Code,
			"exception.Message":    exception.Message,
			"exception.StackTrace": exception.StackTrace}, "Failed to execute clickhouse request:")
	} else {
		log.ErrorWithFields(log.Fields{"error": err}, "Failed to execute clickhouse request:")
	}
}

func (db *Clickhouse) runMigrations(dbname string) bool {
	log.Debug("Checking clickhouse DB schema...")
	cfg := &mm.Config{DatabaseName: dbname}

	driver, err := mm.WithInstance(db.Db, cfg)
	if err != nil {
		log.ErrorWithFields(log.Fields{"error": err}, "Migrations: driver failed")

		return false
	}

	src, _ := source.Open("stub://")
	src.(*stub.Stub).Migrations = db.migrations
	m, err := migrate.NewWithInstance("", src, "", driver)
	m.Log = storage.MigrateLogger{}

	if err != nil {
		log.ErrorWithFields(log.Fields{"error": err}, "Migrations: instance failed")

		return false
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.ErrorWithFields(log.Fields{"error": err}, "Migrations: process failed")

		return false
	}

	return true
}

func (db *Clickhouse) getDsnQueryParam(name string) string {
	url, err := url.Parse(db.dsn.String())
	if err != nil {
		return ""
	}

	return url.Query().Get(name)
}

func (db *Clickhouse) Stats(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
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
	}

	return metrics, nil
}
