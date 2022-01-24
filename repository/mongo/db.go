package mongo

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"github.com/golang-migrate/migrate/v4"
	mm "github.com/golang-migrate/migrate/v4/database/mongodb"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type EventHandler = func()

type MongoDb struct {
	Client            *mongo.Client
	Db                *mongo.Database
	onConnectHandlers []EventHandler
	isReady           int32

	doOnce       sync.Once
	migrationSrc source.Driver
}

//NewMongoDbConnection creates new connection to mongo db
func NewMongoDbConnection() *MongoDb {
	return &MongoDb{}
}

const timeout = 5 * time.Second

//SetMigrationSrc should be called before Connect
func (db *MongoDb) SetMigrationSrc(src source.Driver) {
	db.migrationSrc = src
}

//InitDB initializes db connection
func (db *MongoDb) Connect(server config.ConnectionString, database string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Set client options
	monitor := &event.PoolMonitor{Event: db.poolEventMonitor}
	clientOptions := options.Client().ApplyURI(server.String()).SetPoolMonitor(monitor).SetServerSelectionTimeout(5 * time.Second)
	var err error
	db.Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.FatalWithFields(log.Fields{"server": &server, "error": err}, "MongoDB connection failed")
	}
	db.Db = db.Client.Database(database)

	// Check the connection
	err = db.Client.Ping(ctx, nil)
	if err != nil {
		log.ErrorWithFields(log.Fields{"server": &server, "error": err}, "MongoDB ping failed")
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
	config := &mm.Config{DatabaseName: dbname}
	driver, err := mm.WithInstance(db.Client, config)
	if err != nil {
		log.ErrorWithFields(log.Fields{"error": err}, "Migrations: driver failed")
		return false
	}

	m, err := migrate.NewWithInstance("", db.migrationSrc, "", driver)
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
