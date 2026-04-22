package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/storage"
	"github.com/golang-migrate/migrate/v4"
	mm "github.com/golang-migrate/migrate/v4/database/mongodb"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/stub"
	"github.com/pkg/errors"
	mongoV1 "go.mongodb.org/mongo-driver/mongo"
	mongoV1Options "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
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
	clientOptions := options.Client().ApplyURI(db.config.Server.String()).
		SetPoolMonitor(monitor).
		SetServerSelectionTimeout(timeout)

	var err error
	db.Client, err = mongo.Connect(clientOptions)
	if err != nil {
		log.PanicWithFields(log.Fields{"server": &db.config.Server, "error": err}, "MongoDB v2 connection failed")
	}
	db.Db = db.Client.Database(db.config.DbName)

	if err = db.Client.Ping(ctx, readpref.Primary()); err != nil {
		log.PanicWithFields(log.Fields{"server": &db.config.Server, "error": err}, "MongoDB v2 ping failed")
	}

	<-db.migrationReadyC
}

func (db *Storage) Disconnect(ctx context.Context) error {
	if err := db.Client.Disconnect(ctx); err != nil {
		return errors.Wrap(err, "MongoDB v2 disconnect failed")
	}
	log.Info("MongoDB v2 disconnected.")
	return nil
}

func (db *Storage) poolEventMonitor(ev *event.PoolEvent) {
	switch ev.Type {
	case event.ConnectionReady:
		// run migrations only once
		db.dbOnce.Do(func() {
			go func() {
				log.InfoWithFields(
					log.Fields{"server": ev.Address, "connectionID": ev.ConnectionID},
					"MongoDB v2 connected!")
				// then run migrations
				if err := db.runMigrations(db.Db.Name()); err != nil {
					log.Errorf("Mongo v2 migrations: %s", err)
				}
				close(db.migrationReadyC)
			}()
		})
	case event.ConnectionClosed:
		if ev.Reason != "stale" {
			log.InfoWithFields(
				log.Fields{"server": ev.Address, "connectionID": ev.ConnectionID, "reason": ev.Reason},
				"MongoDB v2 connection closed!")
		}
	default:
	}
}

func (db *Storage) runMigrations(dbname string) error {
	if _, ok := db.migrations.First(); !ok {
		return nil
	}

	// golang-migrate's MongoDB driver currently depends on the v1 mongo-driver.
	// Use a temporary v1 client for running migrations while the main storage uses v2.
	v1Client, _, err := db.legacyClient()
	if err != nil {
		return fmt.Errorf("migration client failed: %w", err)
	}
	defer v1Client.Disconnect(context.Background())

	config := &mm.Config{DatabaseName: dbname, MigrationsCollection: db.config.MigrationCollection}
	driver, err := mm.WithInstance(v1Client, config)
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

func (db *Storage) legacyClient() (*mongoV1.Client, *mongoV1.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cli, err := mongoV1.Connect(ctx, mongoV1Options.Client().ApplyURI(db.config.Server.String()))
	if err != nil {
		return nil, nil, err
	}
	dbV1 := cli.Database(db.config.DbName)
	return cli, dbV1, nil
}

func (db *Storage) Stats(ctx context.Context) (interface{}, error) {
	collections, err := db.Db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	stats := make([]interface{}, 0, len(collections)+1)
	res := db.Db.RunCommand(ctx, bson.M{"dbStats": 1})
	if err := res.Err(); err != nil {
		return nil, err
	}

	dbStat := make(map[string]interface{})
	if err := res.Decode(&dbStat); err != nil {
		return nil, err
	}
	stats = append(stats, dbStat)

	for _, collection := range collections {
		res := db.Db.RunCommand(ctx, bson.M{"collStats": collection})
		if err := res.Err(); err != nil {
			return nil, err
		}
		colStat := make(map[string]interface{})
		if err := res.Decode(&colStat); err != nil {
			return nil, err
		}
		stats = append(stats, colStat)
	}
	return stats, err
}
