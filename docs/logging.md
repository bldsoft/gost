# Logging

gost wraps zerolog in a global `log.Logger` plus package-level helpers, and ties request-scoped
loggers to `context.Context` so every line of a request carries its `req_id`.

## Initialisation

<!-- source: examples/notes/cmd/notes/main.go -->
```go
log.InitLogger(&cfg.Log, cfg.Server.LogExporterConfig())
log.Infof("%s v%s", config.ServiceName, version.Version)
log.Logf("ENVIRONMENT:\n***\n%s***", gost.FormatEnv(&cfg))
```

Until `InitLogger` runs, `log.Logger` writes plain JSON to stderr, so logging during
`ReadConfig` still works. After it, output goes to the console (colourised when
`LOG_COLOR_ENABLED=true`), to `LOG_FILE` if set (lumberjack rotation, 50 MB, 3 backups, 30 days),
and to any registered exporter. Field names are short: `t` (unix micro), `l` (level), `m` (message).

| Variable | Default | Meaning |
|---|---|---|
| `LOG_LEVEL` | `info` | zerolog level: trace, debug, info, warn, error |
| `LOG_COLOR_ENABLED` | `true` | Colourised console output; set `false` in containers |
| `LOG_FILE` | empty | Optional rotating log file (hidden from `config.md`) |

`log.SetLogLevel("debug")` changes the level at runtime; services wire it to a feature flag.

## Writing log lines

Package functions exist for every level in four shapes: `Info`, `Infof`, `InfoWithFields`,
`InfofWithFields`, and likewise for `Trace`, `Debug`, `Warn`, `Error`, `Fatal`, `Panic`, `Log`.
`DebugOrError*` and `InfoOrError*` take an `error` first and pick the level from whether it is nil.

```go
log.ErrorWithFields(log.Fields{"error": err}, "Failed to purge soft-deleted notes")
```

`log.Fields` is `map[string]any`. Put the error under `"error"` and keep messages constant so
they can be searched; variable data goes into fields.

## Request-scoped logging

`UseDefaultMiddlewares()` installs gost's request logger. It logs `METHOD REQUEST` with uri,
proto, method, remote address and headers, then `RESPONSE` with status, size and duration, and
it stores a logger carrying `req_id` in the request context. Inside handlers, services and
repositories always use:

```go
log.FromContext(ctx).ErrorWithFields(log.Fields{"error": err}, "Note request failed")
```

`FromContext` falls back to the global logger when there is no request, so it is safe in
background jobs too. To enrich the logger for the rest of a request:

```go
ctx := context.WithValue(r.Context(), log.LoggerCtxKey, log.FromContext(r.Context()).WithFields(log.Fields{"db": name}))
```

Paths listed in `log.NotLoggedEndpoints` are not logged by the request logger; in-house
discovery adds its transport endpoint there automatically. Add noisy health or metrics paths in
`BuildRoutes` or an `init()`.

### Formatters

`server.DefaultLogger` is `log.DefaultRequestLogger()`. Alternatives:

- `log.NewRequestLogger(formatter)` with your own `LogFormatter`.
- `log.MultiLogger(log.NewDefaultFormatter(), log.NewExportFormatter(exporter, instance))` to
  also export per-request records (for example into ClickHouse).
- `log.ChanRequestLogger(ch, instance)` to receive `log.RequestInfo` values on a channel.

`controller.ResponseError` with a 5xx code attaches the error text to the exported request
record through `log.AsResponseWriterLogErr`.

## Exporting logs

`log.Logger.AddExporter(exporter, isOn)` sends every record to a `log.LogExporter`. The
ClickHouse implementation also serves log queries:

```go
clickhouseDB := clickhouse.NewStorage(cfg.Clickhouse)
storage.DBConnectAsync(&wg, clickhouseDB.Connect, -1, time.Second)
wg.Wait()
logExporter := clickhouse.NewLogExporter(clickhouseDB, cfg.LogExport)   // CLICKHOUSE_LOG_EXPORT_*
log.Logger.AddExporter(logExporter, features.LogExportOn)                // *feature.Bool or nil
asyncRunners = append(asyncRunners, logExporter)                         // it is an AsyncRunner
```

`cfg.Server.LogExporterConfig()` supplies the service and instance names that are stored with
each record. See [Observability](observability.md) for the query side.

## Testing with logs

Tests do not need `InitLogger`. The default stderr JSON is fine, and `log.SetLogLevel("")`
silences everything when output is noise.
