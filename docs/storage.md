# Storage

gost ships MongoDB (primary store), ClickHouse (analytics, statistics, log export), several
caches, and two distributed lock implementations. All of them plug into the same patterns:
a `Storage` that implements `storage.IStorage`, connected with `storage.DBConnectAsync`,
disconnected by an `AsyncRunner`, and repositories built on top.

## MongoDB

### Configuration and connection

`mongo.Config` reads `DATABASE_SERVER` (default `mongodb://localhost:27017`), `DATABASE_DBNAME`
(default `test`; override in your `SetDefaults`) and `DATABASE_MIGRATION_COLLECTION`.

<!-- source: examples/notes/pkg/microservice/microservice.go -->
```go
db := mongo.NewStorage(cfg.Mongo)
addMigrations(db)                       // 1. register migrations

func (m *Microservice) Initialize() {   // 2. connect with retries
	var wg sync.WaitGroup
	storage.DBConnectAsync(&wg, m.db.Connect, -1, time.Second)
	wg.Wait()
}
                                        // 3. disconnect last, from an AsyncRunner
server.NewAsyncJobChain(server.NewAsyncJob(nil, m.db.Disconnect), server.NewAsyncJobGroup(...))
```

`Storage.Connect()` **panics** when the server is unreachable and then blocks until migrations
have run. `storage.DBConnectAsync(wg, connect, attempts, sleep)` recovers the panic and retries;
`attempts < 0` retries forever. Use `wg.Wait()` to block startup, or let it run in the background
if the service can serve without the database.

Change streams, transactions, the changelog and watched repositories need a **replica set**. Run
`mongo:8 --replSet rs0` locally and `rs.initiate(...)` once; the CI template does the same.

### Migrations

`db.AddMigration(version, up, down)` takes golang-migrate MongoDB command documents as JSON
arrays. Register them **before** `Connect`; they run once, in version order, and are recorded in
`DATABASE_MIGRATION_COLLECTION` (`schema_migrations`).

<!-- source: examples/notes/pkg/microservice/migrations.go -->
```go
const (
	notesTitleIndexUp = `[{
	"createIndexes": "%s",
	"indexes": [{
		"key": {"title": 1},
		"name": "title_asc"
	}]
}]`
	notesTitleIndexDown = `[{
	"dropIndexes": "%s",
	"index": "title_asc"
}]`
)

func addMigrations(db *mongo.Storage) {
	db.AddMigration(1,
		fmt.Sprintf(notesTitleIndexUp, repository.CollectionName),
		fmt.Sprintf(notesTitleIndexDown, repository.CollectionName))
}
```

Feature flag definitions are also migrations: `feature.AddFeatureMigration(db, version, ...)`.
Keep version numbers strictly increasing and never edit a migration that has shipped.

### Entities

Embed the gost base types with `bson:",inline"`:

<!-- source: examples/notes/pkg/entity/note.go -->
```go
type Note struct {
	mongo.EntityID        `bson:",inline"`   // ID bson.ObjectID `json:"id" bson:"_id"`; RawID/StringID/SetIDFromString/GenerateID
	mongo.EntityTimeStamp `bson:",inline"`   // createTime, updateTime, createUserId, updateUserId; filled by the repository
	mongo.EntityArchived  `bson:",inline"`   // deleteTime, archived; soft delete

	Title string `json:"title" bson:"title"`
	Body  string `json:"body,omitempty" bson:"body,omitempty"`
}
```

Create and update user IDs are taken from the context when it carries a user:
`ctx = mongo.WithUserEntry(ctx, user)`.

### Repositories

`mongo.NewRepository[T](db, "collection")` returns `mongo.Repository[T, *T]`, which implements
the storage-agnostic `repository.Repository[T, *T]` plus `Collection()`, `WithTransaction` and
`AggregateOne`. Embed it and add domain queries:

<!-- source: examples/notes/pkg/repository/note_repository.go -->
```go
type NoteRepository struct {
	mongo.Repository[entity.Note, *entity.Note]
}

func NewNoteRepository(db *mongo.Storage) *NoteRepository {
	return &NoteRepository{Repository: mongo.NewRepository[entity.Note](db, CollectionName)}
}

func (r *NoteRepository) PurgeDeletedBefore(ctx context.Context, before time.Time) error {
	filter := bson.M{mongo.BsonFieldNameDeleteTime: bson.M{"$lt": before}}

	// Without an explicit Archived:false option DeleteMany only marks documents as archived.
	return r.DeleteMany(ctx, filter, &repository.QueryOptions{Archived: false})
}
```

Method families on `repository.Repository`: `FindOne`, `FindByID`, `FindByIDs`,
`FindByStringIDs`, `Find`, `Count`, `GetAll`; `Insert`, `InsertMany`, `Update` (full replace by
`_id`, `ErrNotFound` when missing), `UpdateMany`, `UpdateOne`, `UpdateAndGetByID`, `Upsert*`,
`Replace*`, `InsertOrReplace*`; `Delete`, `DeleteMany`, `FindOneAndDelete`.

IDs may be passed as `string`, `bson.ObjectID` or any `repository.IEntityID`; a malformed string
simply matches nothing (`ErrNotFound`). Errors are normalised by `mongo.WrapErr`:
`repository.ErrNotFound`, `repository.ErrAlreadyExists` (duplicate key).

### Query options and soft delete

`repository.QueryOptions{Archived, Fields, Sort, Filter, Limit, Offset}` is the last, optional
argument of most methods.

- **Reads hide archived documents** unless `Archived: true`.
- **`Delete` and `DeleteMany` soft-delete** (set `deleteTime` and `archived`) unless you pass an
  option with `Archived: false`, which performs a real delete. This is the single most surprising
  rule in gost; the example's purger shows the hard-delete path.
- `Filter` may be a struct: non-zero fields are matched by `bson` tag, slices become `$in`
  (`mongo.RecursiveParse`). `controller.QueryOptionFromRequest[F]` builds one from a query string.
- `Sort: repository.Sort().Desc("createTime")`, `Limit`, `Offset` map to the driver options.

### Cached and watched repositories

- `mongo.NewWatchedRepository[T](db, coll, watchers...)` publishes create/update/delete events to
  `repository.Watcher` values using a change stream with a polling fallback. Call
  `EnablePreAndPostImages(ctx)` so delete events carry the full document.
- `mongo.NewCachedRepository[T](db, coll, localCache, mongo.CachedRepositoryOptions{WarmUp: true, CacheKeyPrefix: "x"})`
  is a read-through cache (`FindByID`, `FindByIDs`) kept coherent by the watcher. `localCache`
  is any `cache.ILocalCacheRepository`, for example `bigcache.NewRepository(jsonConfig)`.
- `changelog_mongo.WrapRepository(repo, changeLogRepo)` records every write in a `change_log`
  collection inside a transaction; see [Observability](observability.md#changelog).

### Mocks

`mongo/mocks` contains mockery mocks of `mongo.Repository` and friends. `mongo/mock` is an old
gomock file that no longer matches the interface; do not use it. Services normally mock their
own narrow repository interface instead (see [Testing](testing.md)).

### Transactions

```go
_, err := repo.WithTransaction(ctx, func(ctx context.Context) (any, error) {
	if err := repo.Insert(ctx, a); err != nil { return nil, err }
	return nil, other.Insert(ctx, b)
})
```

## ClickHouse

```go
chCfg := clickhouse.Config{Dsn: cfg.ClickhouseDSN}        // or embed clickhouse.Config `mapstructure:"CLICKHOUSE"`
chDB := clickhouse.NewStorage(cfg.Clickhouse)             // optional .InCluster("cluster")
chDB.AddMigration(1, createTableSQL, dropTableSQL)        // plain SQL, multi-statement allowed
storage.DBConnectAsync(&wg, chDB.Connect, -1, time.Second)
wg.Wait()
if chDB.IsReady() {
	chDB.RunMigrations()                                  // explicit, unlike Mongo
}
```

Differences from Mongo: `Connect()` logs instead of panicking and sets `IsReady()`; migrations
run only when you call `RunMigrations()`; `Validate()` accepts an empty DSN so ClickHouse can be
optional, and converts v1 `tcp://...?database=x` DSNs to the v2 format.

Building blocks:

- `clickhouse.NewBaseRepository(chDB)` with `RunSelect(ctx, squirrel.SelectBuilder)`, `LogQuery`
  and `GetChartValues(ctx, subQuery, from, to, step)` for time series.
- `clickhouse.NewExporter[T](chDB, clickhouse.ExporterConfig{TableName, BufferedExporterConfig})`
  is a batched, non-blocking inserter and an `AsyncRunner`. Columns come from `ch:"name"` struct
  tags. `BufferedExporterConfig{MaxFlushInterval, MaxBatchSize, ChanBufSize, PreserveOld, Logger}`
  lives in `utils/exporter`.
- `clickhouse.NewLogExporter(chDB, cfg.LogExport)` implements `log.LogExporter`; see
  [Logging](logging.md#exporting-logs).
- `stat.NewClickHouseCollector(chDB)` for `/stat`.

## Caches

Two generations exist. New code should use **`cache/v2`** for distributed caches and the v1
local backends for in-process caches.

| Need | Use |
|---|---|
| In-process byte cache | `cache/bigcache.NewRepository(jsonConfig)` or `cache/ristretto.NewRepository(jsonConfig)` (implement `cache.ILocalCacheRepository`) |
| In-process with TTL | `cache.NewExpiringRepository(local)` |
| Typed values | `cache.Typed[T](expiring)` (gob by default, `.WithJSONMarshaler()`) |
| Distributed | `cache/v2/memcached.NewMemcacheRepository(memcached.NewStorage(servers, cfg), ttl)` or `cache/v2/aerospike.NewRepository(storage, ttl, set)` |
| Mongo-backed TTL cache | `cache/mongo.NewMongoCache(db)` |
| Distributed mutex on a cache | `cache.NewDistrMutex(distr, key, unlockTime)` (v1 and v2) |

`cache/v2/aerospike.Config` has `SetDefaults` but no `Validate`, so it is not an `IConfig` and
gost does not apply its defaults automatically; call them yourself or wrap it.

## Distributed locks

- `distlock.NewMongoDistLock(db, lockID, ttl)` returns a `DistrMutex` (`Lock(ctx)`, `TryLock`,
  `Unlock`, `Quit()`), auto-renewed every `ttl/2`, stored in collection `distr_mutex`.
- `mongo-lock` (package `mongolock`) is the lower-level library: `NewClient(coll)`,
  `CreateIndexes`, `XLock`, `SLock`, `Unlock`, `Renew`, `Status`, `NewPurger(client).Purge(ctx)`.
