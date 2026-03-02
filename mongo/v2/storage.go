package v2

import (
	"context"
	"sync"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/golang-migrate/migrate/v4/source"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Storage struct {
	config Config

	Client *mongo.Client
	Db     *mongo.Database

	dbOnce          sync.Once
	migrations      *source.Migrations
	migrationReadyC chan struct{}
}

func NewStorage(config Config) *Storage {
	return &Storage{config: config, migrations: source.NewMigrations(), migrationReadyC: make(chan struct{})}
}

func (db *Storage) AddMigration(version uint, migrationUp, migrationDown string) {
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Up, Identifier: migrationUp})
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Down, Identifier: migrationDown})
}

const timeout = 5 * time.Second

func (db *Storage) Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	monitor := &event.PoolMonitor{Event: db.poolEventMonitor}
	clientOpts := options.Client().ApplyURI(db.config.Server.String()).SetPoolMonitor(monitor).SetServerSelectionTimeout(timeout)
	var err error
	db.Client, err = mongo.Connect(ctx, clientOpts)
}

func (db *Storage) poolEventMonitor(e *event.PoolEvent) {
	switch e.Type {
	case event.ConnectionReady:
		db.dbOnce.Do(func() {
			go func() {
				log.InfoWithFields(log.Fields{"server": e.Address, "connectionID": e.ConnectionID}, "MongoDB connected!")

				if err := db.runMigrations(db.Db.Name()); err != nil {
					log.Errorf("Mongo migrations: %s", err)
				}
				close(db.migrationReadyC)
			}()
		})

	case event.ConnectionClosed:
		if e.Reason != "stale" {
			log.InfoWithFields(
				log.Fields{"server": e.Address, "connectionID": e.ConnectionID, "reason": e.Reason},
				"MongoDB connection closed!")
		}
	}
}

func (db *Storage) runMigrations(dbname string) error {
	if _, ok := db.migrations.First(); !ok {
		return nil
	}

	config := &mm.Config{DatabaseName: dbname, MigrationsCollection: db.config.MigrationCollection}
	driver, err := mm.WithInstance(db.Client, config)
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
