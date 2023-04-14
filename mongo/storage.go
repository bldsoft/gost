package mongo

import (
	"context"
	"sync"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/storage"
	"github.com/golang-migrate/migrate/v4"
	mm "github.com/golang-migrate/migrate/v4/database/mongodb"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/stub"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventHandler = func()

type Storage struct {
	config Config

	Client *mongo.Client
	Db     *mongo.Database

	doOnce          sync.Once
	migrations      *source.Migrations
	migrationReadyC chan struct{}
}

// NewStorage ...
func NewStorage(config Config) *Storage {
	return &Storage{config: config, migrations: source.NewMigrations(), migrationReadyC: make(chan struct{})}
}

// AddMigration adds a migration. All migrations should be added before db.Connect
func (db *Storage) AddMigration(version uint, migrationUp, migrationDown string) {
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Up, Identifier: migrationUp})
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Down, Identifier: migrationDown})
}

const timeout = 5 * time.Second

// Connect initializes db connection
func (db *Storage) Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set client options
	monitor := &event.PoolMonitor{Event: db.poolEventMonitor}
	clientOptions := options.Client().ApplyURI(db.config.Server.String()).SetPoolMonitor(monitor).SetServerSelectionTimeout(5 * time.Second)
	var err error
	db.Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.PanicWithFields(log.Fields{"server": &db.config.Server, "error": err}, "MongoDB connection failed")
	}
	db.Db = db.Client.Database(db.config.DbName)

	// Check the connection
	err = db.Client.Ping(ctx, nil)
	if err != nil {
		log.PanicWithFields(log.Fields{"server": &db.config.Server, "error": err}, "MongoDB ping failed")
	}

	<-db.migrationReadyC
}

// Disconnect closes db connection
func (db *Storage) Disconnect(ctx context.Context) error {
	err := db.Client.Disconnect(ctx)
	if err != nil {
		return errors.Wrap(err, "MongoDB disconnect failed")
	}
	log.Info("MongoDB disconnected.")
	return nil
}

func (db *Storage) poolEventMonitor(ev *event.PoolEvent) {
	switch ev.Type {
	case event.ConnectionReady:
		//run migrations only once
		db.doOnce.Do(func() {
			go func() {
				log.InfoWithFields(
					log.Fields{"server": ev.Address, "connectionID": ev.ConnectionID},
					"MongoDB connected!")
				//then run migrations
				db.runMigrations(db.Db.Name())
				close(db.migrationReadyC)
			}()
		})
	case event.ConnectionClosed:
		if ev.Reason != "stale" {
			log.InfoWithFields(
				log.Fields{"server": ev.Address, "connectionID": ev.ConnectionID, "reason": ev.Reason},
				"MongoDB connection closed!")
		}
	default:
		// log.Debugf("MogoDB event: %v", *ev)
	}
}

func (db *Storage) IsReady() bool {
	return true
}
func (db *Storage) runMigrations(dbname string) bool {
	log.Debug("Checking DB schema...")

	if _, ok := db.migrations.First(); !ok {
		return true
	}

	config := &mm.Config{DatabaseName: dbname}
	driver, err := mm.WithInstance(db.Client, config)
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
