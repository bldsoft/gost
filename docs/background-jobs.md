# Background jobs and shutdown

Everything that outlives a request is a `server.AsyncRunner`. The server starts all runners
after routes are built and stops them, within 5 seconds, on SIGINT/SIGTERM/SIGQUIT. Services
never write their own signal handling.

```go
type AsyncRunner interface {
	Run() error                      // blocks for the lifetime of the job
	Stop(ctx context.Context) error  // must return once the job has exited, or when ctx expires
}
```

## Writing a runner

Use `server.NewContextAsyncRunner` for anything shaped like a loop. `Stop` cancels the context
and waits for `Run` to return:

<!-- source: examples/notes/pkg/service/purger.go -->
```go
func NewPurger(notes NoteRepository, interval, after time.Duration) server.AsyncRunner {
	return server.NewContextAsyncRunner(func(ctx context.Context) error {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-ticker.C:
				if err := notes.PurgeDeletedBefore(ctx, time.Now().Add(-after)); err != nil {
					log.ErrorWithFields(log.Fields{"error": err}, "Failed to purge soft-deleted notes")
				}
			}
		}
	})
}
```

Other constructors:

| Constructor | Use |
|---|---|
| `server.NewContextAsyncRunner(func(ctx) error)` | Loops and pollers; recommended default |
| `server.NewAsyncJob(run func() error, stop func(ctx) error)` | Adapt a type with its own start/stop; either may be nil |
| `server.NewAsyncJobGroup(r...)` | Run and stop in parallel; recovers and logs panics in `Run` |
| `server.NewAsyncJobChain(r...)` | Start in order, **stop in reverse order** |
| `server.NewAsyncJobManager(r...)` | What the server uses internally |

gost types that already are runners: `discovery.Discovery`, `clickhouse.Exporter`,
`clickhouse.ClickHouseLogExporter`, `exporter.BufferedExporter`, `mongo.Watcher` based
repositories, `alert.Manager` (via `Run(ctx)` with a context runner).

## Registering runners

Return them from the microservice:

<!-- source: examples/notes/pkg/microservice/microservice.go -->
```go
func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return []server.AsyncRunner{
		server.NewAsyncJobChain(
			server.NewAsyncJob(nil, m.db.Disconnect),
			server.NewAsyncJobGroup(m.purger, m.discovery),
		),
	}
}
```

Reading the chain: on start, the no-op job returns immediately and the group starts the purger
and discovery in parallel. On stop, the chain stops the group first and only then calls
`db.Disconnect`, so no job sees a closed connection. Use this shape whenever runners depend on a
storage; real services call it a "db job group".

Runners created in `main` (for example a second listener) go through
`server.AddAsyncRunners(...)`.

## Shutdown rules

- `Stop` must honour its context; the server's budget is 5 seconds for all runners together.
- `Run` returning early is logged (`<Type> job ended`) but does not stop the service. If a job
  must stay alive, loop and log inside it.
- `GetAsyncRunners` is called after `BuildRoutes`. Work that must be done before traffic (DB
  connection, warm-up) belongs in `Initialize()`.
- Keep runners that bind sockets or need the network **out** of `GetAsyncRunners` when tests
  construct the microservice; expose them through a separate method and add them in `main`.

## Scheduling helpers

- `utils/worker_pool.NewIntervalRunner(workers)` runs many `(id, interval, func(ctx))` jobs on a
  pool; wrap its `Run(ctx)` in a context runner.
- `utils/worker_pool.WorkerPool` and `NewWorkerPoolWithFeature(*feature.Int, default)` for
  fan-out with a runtime adjustable worker count.
- `utils/errgroup.Group` collects errors and re-panics captured panics on `Wait`.
- `utils/timedqueue` is a time-ordered queue used by the alert manager.
