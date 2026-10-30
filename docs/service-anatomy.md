# Service anatomy

A gost service is a thin composition of gost packages around your domain code, organised by the
Clean Architecture principles: entities at the centre, use cases in services, adapters (HTTP,
storage) on the outside, and dependencies pointing inwards through interfaces the inner layer
declares. This page fixes the directory layout, the dependency direction between layers, and
what happens between `main()` and the first request.

## Layout

```
cmd/<svc>/main.go          entry point: ReadConfig → InitLogger → NewServer(...).Start()
pkg/config/                Config struct composed from gost configs, SetDefaults, Validate, golden test for config.md
pkg/entity/                domain types; imports gost base types only, never another layer
pkg/repository/            storage adapters (Mongo, ClickHouse, HTTP upstreams); never imports service
pkg/service/               business logic; interface.go declares the repositories it consumes
pkg/controller/            interface.go declares the services controllers consume
pkg/controller/rest/       HTTP handlers; parse → call service → respond; no business logic
pkg/microservice/          composition root: wiring, migrations, routes, async runners
VERSION                    single source of the version, injected through ldflags
config.md                  generated env var table (golden file)
Makefile, Dockerfile, .golangci.yml, .mockery.yml, AGENTS.md, .github/workflows/ci.yml
```

Older services use `internal/` instead of `pkg/`, or group by domain (`internal/score/{controller,service}.go`).
Both work; what matters is the dependency direction below.

## Dependency direction

```
controller ──▶ service ──▶ repository ──▶ entity
     │             │             │
     └─────────────┴─────────────┴──────▶ gost packages
```

- Each layer **declares the interfaces it consumes**: `pkg/controller/interface.go` holds the
  service interfaces, `pkg/service/interface.go` holds the repository interfaces. Implementations
  live one layer down and are checked at compile time in the composition root:

  <!-- source: examples/notes/pkg/microservice/microservice.go -->
  ```go
  var (
  	_ service.NoteRepository           = (*repository.NoteRepository)(nil)
  	_ controller.NoteService           = (*service.NoteService)(nil)
  	_ server.InitializableMicroservice = (*Microservice)(nil)
  )
  ```
- Domain errors are sentinel values in `pkg/entity` (for example `entity.ErrInvalidNote`) so that
  controllers can map them to status codes without importing the service package.
- Repository errors come from gost: `repository.ErrNotFound`, `repository.ErrAlreadyExists`.
- Dependency injection is manual, in constructors. No fx or wire.

## The composition root

`pkg/microservice` implements `server.IMicroservice` and, when it needs a database,
`server.InitializableMicroservice`:

<!-- source: examples/notes/pkg/microservice/microservice.go -->
```go
func New(cfg config.Config) *Microservice {
	db := mongo.NewStorage(cfg.Mongo)
	addMigrations(db)                                   // before Connect, nothing touches the network here

	notesRepository := repository.NewNoteRepository(db)
	notesService := service.NewNoteService(notesRepository)

	return &Microservice{
		cfg:       cfg,
		db:        db,
		discovery: common.NewDiscovery(cfg.Server, cfg.Discovery),
		purger:    service.NewPurger(notesRepository, cfg.PurgeInterval, cfg.PurgeAfter),
		notes:     rest.NewNoteController(notesService),
	}
}

func (m *Microservice) Initialize() {                   // called by the server right before BuildRoutes
	var wg sync.WaitGroup
	storage.DBConnectAsync(&wg, m.db.Connect, -1, time.Second)
	wg.Wait()
}

func (m *Microservice) BuildRoutes(r server.Router) {
	if d, ok := m.discovery.(*inhouse.Discovery); ok {
		d.Mount(r)
	}

	r.Route(BasePath, func(r server.Router) {
		r.Get("/ping", gost.GetPingHandler)
		r.Get("/version", gost.GetVersionHandler)
		r.With(auth.ApiKeyMiddleware(APIKeyHeader, m.cfg.AdminAPIKey.String())).
			Get("/env", gost.GetEnvHandler(&m.cfg, nil))

		r.Route("/api/v1/notes", m.notes.Mount)
	})
}

func (m *Microservice) GetAsyncRunners() []server.AsyncRunner {
	return []server.AsyncRunner{
		server.NewAsyncJobChain(
			server.NewAsyncJob(nil, m.db.Disconnect),
			server.NewAsyncJobGroup(m.purger, m.discovery),
		),
	}
}
```

Rules that keep it testable:

- `New` must not open sockets or connect to anything. Tests call `New(cfg)` and `BuildRoutes`
  on a bare `chi.NewRouter()`.
- Network connections happen in `Initialize()`, which only the real server calls.
- Anything long lived is a `server.AsyncRunner` returned from `GetAsyncRunners()`. See
  [Background jobs](background-jobs.md).

## Startup sequence

What `server.Server.Start()` does, in order (`server/server.go`):

1. With `WithHealthProbes(true)`: install a probes-only router on which `/probes/ready` answers
   503, then build the application router in a goroutine. Without probes: build it synchronously.
   With `PROBES_BIND_ADDRESS` or an `InternalServer` the probes are mounted on the internal
   router instead and neither main router carries them.
2. Building the application router: for each microservice, call `Initialize()` if implemented,
   then `BuildRoutes(r)` inside its own `chi` group. The common middlewares, `/debug/pprof` and
   `/probes/*` (ready = 200) are mounted on the same router; with an internal listener the
   ready flag flips there once this step completes.
3. Collect `GetAsyncRunners()` from all microservices and start them in parallel, after the
   routes exist.
4. Listen on `SERVICE_BIND_ADDRESS` (the TLS address when `TLS_CERTIFICATE_PATH` is set, and
   the internal address when probes are moved) and block until SIGINT, SIGTERM or SIGQUIT.

## Shutdown sequence

On a signal the server calls `http.Server.Shutdown`, stops the internal listener and all runners
with a 5 second budget, then logs "Server gracefully stopped" or the joined errors. Runners stop in parallel
unless you order them: the chain above stops the group (purger, discovery) first and disconnects
MongoDB last, because `AsyncJobChain.Stop` runs in reverse order.

Quirks to know (documented, not fixed, in [server.md](server.md#known-quirks)): the handler is
registered on `http.DefaultServeMux`, so only one `server.Server` can run per process and tests
must never call `Start()`; the HTTP `Shutdown` receives an already cancelled context, so in-flight
requests are not drained.

## Multiple microservices in one process

`server.NewServer(cfg, msA, msB)` accepts several `IMicroservice` values. Each gets its own chi
group, so middlewares added inside one `BuildRoutes` do not leak into the other. All runners are
collected into one manager.
