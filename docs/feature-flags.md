# Feature flags (runtime settings)

Configuration is read once at startup. Values that operators change while the service runs,
such as the log level, GC percent, batch sizes, kill switches, are **features**: typed values with
an integer ID, validated on change, optionally persisted in MongoDB with per-instance overrides,
and exposed through a REST controller.

Two packages cooperate:

- `config/feature` (imported as `feature` in config code): the in-memory typed value.
- `feature`: MongoDB persistence, watcher, service and controller.

## Declaring features

```go
package config

import "github.com/bldsoft/gost/config/feature"

type Features struct {
	LogLevel     *feature.String
	PurgeEnabled *feature.Bool
	BatchSize    *feature.Int
	Timeout      *feature.Duration
}

func (f *Features) Init() {
	f.LogLevel = feature.NewFeature(1, "info").
		SetValidator(func(level string) error { _, err := zerolog.ParseLevel(level); return err }).
		AddOnChangeHandler(func(level string) { _ = log.SetLogLevel(level) })
	f.PurgeEnabled = feature.NewFeature(2, true)
	f.BatchSize = feature.NewFeature(3, 1000).SetValidator(func(n int) error { if n <= 0 { return errors.New("must be positive") }; return nil })
	f.Timeout = feature.NewDuration(4, 30*time.Second)
}
```

Rules learned the hard way:

- **IDs are stable integers.** They are the primary key in the database. Never reuse a removed
  ID; leave a comment or a tombstone constant so nobody does.
- Validators run before the value changes; on-change handlers run after. Keep handlers quick.
- Read with `f.BatchSize.Get()`; set programmatically with `Set(v)` or `SetFromString(s)`.
- `feature.NewCustomFeature(id, value, parse)` accepts any comparable type with your own parser,
  useful to keep old serialized forms readable.
- The global registry `feature.Features.Get(id)` is what the persistence layer uses, so
  features must be created before `feature.NewMongoRepository` loads them.

Expose the current values on `/env` by passing the struct as the second argument of
`controller.GetEnvHandler(&cfg, &features)`.

## Persisting in MongoDB

Register definitions as a migration so names, descriptions and groups land in the `feature`
collection:

```go
feature.AddFeatureMigration(db, 5,
	feature.NewFeature(features.LogLevel, "LogLevel", "Log level: trace|debug|info|warn|error").WithGroups("Logs"),
	feature.NewFeature(features.PurgeEnabled, "PurgeEnabled", "Run the soft-delete purger").WithGroups("Maintenance"),
)
```

Use `feature.NewFeatureMigrator(db, "<svc>_features").AddFeatureMigration(...)` when several
services share one database and need separate collections.

Then wire repository, service and controller after the database is connected:

```go
featureRepo := feature.NewMongoRepository(db, cfg.Server.ServiceInstance)   // loads values, applies per-instance overrides, watches for changes
featureService := feature.NewService(featureRepo)
r.Route("/api/env/feature", feature.NewController(featureService).Mount)    // GET /, GET /{id}, PATCH /{id}; protect with admin auth
```

`feature.CreateController(storage, instanceName)` does the three steps in one call for a
`*mongo.Storage`. Pass the **instance name**, not the database name: per-instance values are
matched on it.

## Operating

- `GET /api/env/feature` lists features with global and per-instance values.
- `PATCH /api/env/feature/{id}` with `{"globalValue": "debug"}` or a `srvValues` entry changes a
  value; the Mongo watcher propagates it to all instances within seconds (polling fallback every
  5 minutes, `mongo.ReserveWatcherInterval`).
- Feature-gated gost behaviour: `log.Logger.AddExporter(exp, features.LogExportOn)`,
  `worker_pool.NewWorkerPoolWithFeature(features.Workers, 4)`.

Common mistakes seen in real services: creating the repository and discarding it (flags then
never update), never registering the feature migration (flags invisible in the UI), and calling
`Features.Init()` twice.
