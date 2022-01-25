package mongo

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"github.com/golang-migrate/migrate/v4"
	mm "github.com/golang-migrate/migrate/v4/database/mongodb"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/stub"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventHandler = func()

type MongoDb struct {
	config Config

	Client            *mongo.Client
	Db                *mongo.Database
	onConnectHandlers []EventHandler
	isReady           int32

	doOnce     sync.Once
	migrations *source.Migrations
}

//NewMongoDbConnection creates new connection to mongo db
func NewMongoDbConnection(config Config) *MongoDb {
	return &MongoDb{config: config, migrations: source.NewMigrations()}
}

//AddMigration adds a migration. All migrations should be added before db.Connect
func (db *MongoDb) AddMigration(version uint, migrationUp, migrationDown string) {
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Up, Identifier: migrationUp})
	db.migrations.Append(&source.Migration{Version: version, Direction: source.Down, Identifier: migrationDown})
}

const timeout = 5 * time.Second

//InitDB initializes db connection
func (db *MongoDb) Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set client options
	monitor := &event.PoolMonitor{Event: db.poolEventMonitor}
	clientOptions := options.Client().ApplyURI(db.config.Server.String()).SetPoolMonitor(monitor).SetServerSelectionTimeout(5 * time.Second)
	var err error
	db.Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.FatalWithFields(log.Fields{"server": &db.config.Server, "error": err}, "MongoDB connection failed")
	}
	db.Db = db.Client.Database(db.config.DBName)

	// Check the connection
	err = db.Client.Ping(ctx, nil)
	if err != nil {
		log.ErrorWithFields(log.Fields{"server": &db.config.Server, "error": err}, "MongoDB ping failed")
	}
}

//DisconnectDB closes db connection
func (db *MongoDb) Disconnect(ctx context.Context) error {
	err := db.Client.Disconnect(ctx)
	if err != nil {
		return errors.Wrap(err, "MongoDB disconnect failed")
	}
	log.Info("MongoDB disconnected.")
	return nil
}

func (db *MongoDb) poolEventMonitor(ev *event.PoolEvent) {
	switch ev.Type {
	case event.ConnectionCreated:
		//run migrations only once
		db.doOnce.Do(func() {
			go func() {
				log.InfoWithFields(
					log.Fields{"server": ev.Address, "connectionID": ev.ConnectionID},
					"MongoDB connected!")
				//let connection to be finished
				time.Sleep(1 * time.Second)
				//then run migrations
				db.runMigrations(db.Db.Name())
				atomic.StoreInt32(&db.isReady, 1)
				db.notifyConnect()
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

func (db *MongoDb) IsReady() bool {
	return atomic.LoadInt32(&db.isReady) == 1
}

func (db *MongoDb) AddOnConnectHandler(handler EventHandler) {
	db.onConnectHandlers = append(db.onConnectHandlers, handler)
}

func (db *MongoDb) notifyConnect() {
	for _, handler := range db.onConnectHandlers {
		handler()
	}
}

func (db *MongoDb) runMigrations(dbname string) bool {
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
	m.Log = repository.MigrateLogger{}

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
