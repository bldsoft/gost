# Observability

What every gost service exposes out of the box, and the optional pieces for logs, statistics
and audit trails.

## Standard endpoints

| Path | Purpose | Auth |
|---|---|---|
| `/<svc>/ping` | Liveness from the outside; returns `pong` | none |
| `/<svc>/version` | `{"version":"1.2.3","branch":"main","commit":"abc1234"}` from the `version` package | none |
| `/<svc>/env` | Effective configuration as YAML (`config.FormatEnv`), plus feature values if passed | API key |
| `/probes/healthy` | Kubernetes liveness: 200 while the process runs | none |
| `/probes/ready` | Kubernetes readiness: 503 until `Initialize()` and `BuildRoutes` finished, then 200 | none |
| `/probes/startup` | Kubernetes startup probe: 200 | none |
| `/debug/pprof/*` | Go profiler, including `goroutineleak` | none; block at the ingress |
| `/<svc>/api/env/feature` | Feature flags (if wired) | admin |
| `/<svc>/stat` | Storage statistics (if wired) | admin |

Kubernetes manifests should point liveness at `/probes/healthy` and readiness at
`/probes/ready`. If you put metrics or probes on a private port instead, pass
`WithHealthProbes(false)` and serve them from your own listener (an `AsyncRunner`).

## Version information

`version.Version`, `version.GitCommit` and `version.GitBranch` are plain variables set at build
time; `version.LongVersion()` returns `Version#Commit`. They feed `/version`, discovery metadata
and the log exporter, so always build with the ldflags from [Build and release](build-release.md).

## Request logs

The default request logger logs every request and response with `req_id`, status, size and
duration (see [Logging](logging.md)). Add noisy paths to `log.NotLoggedEndpoints`.

## Log export to ClickHouse

`clickhouse.NewLogExporter(chDB, cfg.LogExport)` stores every log record in a ClickHouse table
(default `LOG_RECORDS`) and implements `log.LogExporter`, which also offers query methods:
`Logs(ctx, params)`, `LogsMetrics` (time series), `Instances`, `Services`, `ServiceVersions`,
`RequestIDs`, `ChangeTTL(hours)`. Wire a controller around it when operators need log search in
the service UI.

| Variable | Default |
|---|---|
| `CLICKHOUSE_LOG_EXPORT_FLUSH_TIME_MS` | `30000` |
| `CLICKHOUSE_LOG_EXPORT_MAX_BATCH_SIZE` | `10000` |
| `CLICKHOUSE_LOG_EXPORT_TABLE` | `LOG_RECORDS` |

## Statistics endpoint

```go
r.With(adminAuth).Route("/stat", stat.NewController(
	stat.NewMongoCollector(db),
	stat.NewClickHouseCollector(chDB),
	stat.NewMemcachedCollector(memcachedStorage),      // cache/memcached (v1) storage
).Mount)
```

Each collector returns its storage's native statistics (`serverStatus`, `system.metrics`,
memcached stats) under one JSON document.

## Changelog

An audit log of entity changes stored in the `change_log` collection: who, when, which
operation, which entity, the JSON data or a merge-patch diff, and the request ID.

```go
type Customer struct {
	mongo.EntityID        `bson:",inline"`
	changelog.LoggedEntity `bson:",inline"`      // adds changeID
	...
}

changeLog := changelog_mongo.NewChangeLogRepository(db)
customers := changelog_mongo.WrapRepository(mongo.NewRepository[Customer](db, "customer"), changeLog)   // Insert/Update/Delete are logged in a transaction
r.With(adminAuth).Route("/changelog", changelog.NewController(changeLog).Mount)                        // GET / with filters via utils.FromRequest
```

`changelog.AddContextDetail(ctx, "reason", "manual fix")` attaches extra details to the next
record. The user comes from `auth.UserFromContext`, the request ID from chi. Transactions
require a replica set.

## Alerts

The `alert` package evaluates `Source`s on a schedule, passes alerts through middlewares
(deduplication, grouping, metadata, sorting, logging) and delivers them by email, Slack webhook or
generic webhook with retries: `alert.NewManager(alert.Config{WorkerN: 4})`,
`AddProcessor(alert.Processor{ID, Source, Handler})`, `alert.NewNotifyService(cfg, receivers...)`,
`alert.NotifyHandler(service)`. Severity levels: `Low`, `Medium`, `High`, `Critical`.

## Metrics

gost has no Prometheus package. Services that need `/metrics` add `prometheus/client_golang`
directly and usually serve it on a private listener. Keep high-cardinality values such as session
IDs out of labels.
