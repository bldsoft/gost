# Package reference

One row per package. "Entry points" are what a service author calls; see the linked guide for
usage. Paths are relative to the gost module root.

## Core

| Package | Purpose | Entry points | Notes |
|---|---|---|---|
| `config` | Env configuration, defaults, validation, `config.md` | `ReadConfig`, `SetDefaults`, `Validate`, `ReadFromEnv`, `FormatEnv`, `WriteConfigDescription`, `NewMarkdownFormatter`, `NewCsvFormatter`, `Address`, `HttpAddress`, `ConnectionString`, `HiddenString` | [Configuration](configuration.md) |
| `config/feature` | Typed runtime-mutable values | `NewFeature[T]`, `NewDuration`, `NewCustomFeature`, `Feature.Get/Set/SetValidator/AddOnChangeHandler`, `Features.Get(id)` | [Feature flags](feature-flags.md) |
| `server` | HTTP server, microservice contract, async runners | `NewServer`, `Server.UseDefaultMiddlewares/WithHealthProbes/AddAsyncRunners/SetRouterWrapper/Start`, `NewInternalServer`, `InternalServer.Handle/Mount/Listen/Run/Stop/Addr`, `IMicroservice`, `InitializableMicroservice`, `Router`, `AsyncRunner`, `NewContextAsyncRunner`, `NewAsyncJob`, `NewAsyncJobGroup`, `NewAsyncJobChain`, `Config` | [Server](server.md), [Background jobs](background-jobs.md) |
| `server/middleware` | Request ID and real IP | `RealIP`, `RequestID`, `GetRealIP`, `WithRealIP`, `TrimPort` | |
| `log` | zerolog wrapper, request logging, exporters | `InitLogger`, `Logger`, `FromContext`, `Fields`, `Info/Debug/Warn/Error*`, `DefaultRequestLogger`, `NewRequestLogger`, `MultiLogger`, `NotLoggedEndpoints`, `LogExporter`, `SetLogLevel`, `Config` | [Logging](logging.md) |
| `controller` | Handler helpers and standard endpoints | `BaseController` (`ResponseJson`, `ResponseError`, `ResponseOK`, `GetObjectFromBody`), `GetQueryOption[T]`, `ParseQueryOption[T]`, `GetQueryOptionSlice`, `QueryOptionFromRequest[F]`, `GetPingHandler`, `GetVersionHandler`, `GetEnvHandler` | [Server](server.md#writing-handlers) |
| `version` | Build-time version variables | `Version`, `GitCommit`, `GitBranch`, `LongVersion()` | set via ldflags |
| `repository` | Storage-agnostic repository contract | `Repository[T,U]`, `QueryOptions`, `Sort()`, `IEntityID`, `IEntityIDPtr`, `ToRawID`, `ErrNotFound`, `ErrAlreadyExists`, `Watcher`, `WatchedRepository` | [Storage](storage.md) |
| `storage` | Storage contract and connection retry | `IStorage`, `DBConnectAsync`, `MigrateLogger` | |
| `utils` | Small helpers | `FromRequest[T]`, `FromQuery[T]`, `Query[T]`, `Parse[T]`, `HiddenString`, `FullyHidden`, `PartiallyHidden`, `Hostname`, `Set[K]`, `IPTree`, `RandString`, `ErrObjectNotFound`, `Probe` (ffprobe) | |

## Storage

| Package | Purpose | Entry points | Notes |
|---|---|---|---|
| `mongo` | MongoDB storage, migrations, repositories, watchers, caches | `NewStorage`, `Storage.AddMigration/Connect/Disconnect`, `Config`, `EntityID`, `EntityTimeStamp`, `EntityArchived`, `NewRepository[T]`, `Repository`, `NewWatchedRepository`, `NewCachedRepository`, `NewWatcher`, `WithUserEntry`, `WrapErr`, `BsonFieldName*` | needs a replica set for streams and transactions |
| `mongo/mocks` | mockery mocks | `NewRepository[T,U]`, `NewIWatcher` | prefer mocking your own interface |
| `mongo/mock` | legacy gomock | | stale, do not use |
| `mongo-lock` | Distributed lock primitives | `NewClient`, `CreateIndexes`, `XLock`, `SLock`, `Unlock`, `Renew`, `Status`, `NewPurger` | |
| `distlock` | Mutex interface over mongo-lock | `DistrMutex`, `NewMongoDistLock` | auto-renew every ttl/2 |
| `clickhouse` | ClickHouse storage, migrations, exporters, log export | `NewStorage`, `Storage.AddMigration/Connect/RunMigrations/IsReady/InCluster`, `Config`, `NewBaseRepository`, `NewExporter[T]`, `ExporterConfig`, `NewLogExporter`, `LogExporterConfig`, `NewBatch` | `RunMigrations` is explicit |
| `cache` | v1 cache interfaces, typed and expiring wrappers, mutex | `ILocalCacheRepository`, `IExpiringCacheRepository`, `IDistrCacheRepository`, `Typed[T]`, `NewExpiringRepository`, `NewDistrMutex`, `ErrCacheMiss` | |
| `cache/bigcache`, `cache/ristretto` | In-process byte caches | `NewRepository(jsonConfig)` | |
| `cache/memcached` | v1 memcached | `NewStorage`, `NewMemcacheRepository`, `Config` | used by `stat.NewMemcachedCollector` |
| `cache/mongo` | TTL cache in MongoDB | `NewMongoCache` | |
| `cache/v2` | v2 distributed cache interface with `Item` | `IDistrCacheRepository`, `Item`, `WithTTL`, `WithFlags`, `NewDistrMutex` | preferred for new code |
| `cache/v2/memcached`, `cache/v2/aerospike` | v2 backends | `NewStorage`, `NewMemcacheRepository`, `NewRepository`, `Config` | aerospike `Config` lacks `Validate` |
| `entity/stat` | Chart DTOs | `SeriesData`, `SeriesValues`, `SeriesValue` | used by ClickHouse chart helpers |

## Platform features

| Package | Purpose | Entry points | Notes |
|---|---|---|---|
| `discovery` | Discovery interface, events, clients | `Discovery`, `NotifyingDiscovery`, `NewBaseDiscovery`, `NewEventHandler`, `NewHttpClient`, `NewTransport`, `NewResolver`, `DefaultDialer`, `GrpcDialer` | [Discovery](discovery.md) |
| `discovery/common` | Config and factory | `Config`, `NewDiscovery` | `DISCOVERY_TYPE` |
| `discovery/inhouse` | memberlist gossip backend | `NewDiscovery`, `Discovery.Mount`, `Config` | override `SECRET_KEY` |
| `discovery/consul` | Consul backend | `NewDiscovery`, `Config` | |
| `discovery/fake` | In-memory backend | `NewDiscovery`, `AddService` | tests and `none` |
| `feature` | Persisted feature flags | `NewFeature`, `Feature.WithGroups`, `NewMongoRepository`, `NewService`, `NewController`, `CreateController`, `AddFeatureMigration`, `NewFeatureMigrator` | [Feature flags](feature-flags.md) |
| `changelog`, `changelog/mongo` | Entity audit log | `LoggedEntity`, `Record`, `AddContextDetail`, `NewService`, `NewController`, `NewChangeLogRepository`, `WrapRepository`, `NewLoggedRepository` | needs a replica set |
| `auth` | Sessions, users, API key | `ApiKeyMiddleware`, `NewAuthService`, `NewAuthController`, `AuthController.AuthenticateMiddleware/Mount`, `NewUserService`, `NewUserController`, `UserFromContext`, `User`, `Creds`, `BcryptHasher`, `ErrUnauthorized`, `ErrForbidden` | [Auth](auth.md) |
| `auth/store` | Session stores | `NewMongoDBStore`, `NewMongoDBStoreWithConfig`, `NewRedisStore` | |
| `auth/jwt` | JWT middleware | `JwtConfig`, `JwtAuthMiddlewareFromConfig`, `JwtAuthMiddleware`, `SignToken` | |
| `auth/acl` | IP allow and deny lists | `Config`, `MiddlewareFromConfig`, `IpRange` | |
| `stat` | `/stat` endpoint | `NewController`, `NewService`, `NewMongoCollector`, `NewClickHouseCollector`, `NewMemcachedCollector`, `NewAerospikeCollector` | [Observability](observability.md) |
| `alert` | Alert evaluation and notification | `NewManager`, `Processor`, `Source`, `Middleware`, `NewNotifyService`, `NotifyHandler`, `SeverityLevel`; `alert/middleware`, `alert/notify` | |
| `breaker` | Circuit breaker | `NewCircuitBreaker`, `Settings`, `CircuitBreaker.Execute`, `NewTwoStepCircuitBreaker`, `NewClient` (HTTP) | from sony/gobreaker |

## Utilities

| Package | Purpose | Entry points |
|---|---|---|
| `utils/errgroup` | errgroup that joins errors and re-panics | `Group.Go`, `Group.Wait` |
| `utils/exporter` | Generic batched exporters | `Exporter[T]`, `NewBuffered`, `NewBufferedFromExporter`, `BufferedExporterConfig`, `Func`, `Chan`, `Transform` |
| `utils/worker_pool` | Worker pools and interval runner | `WorkerPool`, `NewWorkerPoolWithFeature`, `NewIntervalRunner` |
| `utils/health_check` | HTTP health checker | `NewHealthChecker`, `Check`, `ErrServiceNotAvailable` |
| `utils/poly` | Polymorphic JSON/BSON values with a `type` discriminator | `Poly[T]`, `Register[I]().Type(name, impl)` |
| `utils/ringbuf`, `utils/heap`, `utils/timedqueue`, `utils/seq`, `utils/slices`, `utils/sync`, `utils/json` | Data structures and helpers | see package files |
| `regexes/ua` | User-agent regex dictionaries (YAML/CSV) | data only |
| `tools` | Pins code generators behind `//go:build tools` | |
| `benchmarks` | `main.txt` baseline maintained by CI | |
