# Agent guide: building a service on gost

Concise rules for coding agents. Humans are welcome too. Everything here is implemented in
[`examples/notes`](../examples/notes); when in doubt, open the matching file there and copy its shape.

## Invariants

1. Entry point is `config.ReadConfig` → `log.InitLogger` → `server.NewServer(...).UseDefaultMiddlewares().WithHealthProbes(true).Start()`. No custom signal handling.
2. `Config` embeds `server.Config`, `log.Config`, and the storage/discovery configs you use. Every leaf field has `mapstructure` and `description` tags.
3. `SetDefaults` sets `c.Server.ServiceName`. Never call nested `SetDefaults`/`Validate`; gost does.
4. Secrets are `config.HiddenString`, have no default, and `Validate` rejects empty values.
5. `config.md` is generated (golden test); edit tags, not the table.
6. `New(cfg)` is network-free. Connect in `Initialize()` via `storage.DBConnectAsync`. Register `db.AddMigration` before that.
7. Layers: `controller → service → repository → entity`. The consumer declares the interface (`pkg/controller/interface.go`, `pkg/service/interface.go`). Conformance via `var _ I = (*T)(nil)` in `pkg/microservice`.
8. Anything long-lived is a `server.AsyncRunner`; return it from `GetAsyncRunners()`, ordered with `NewAsyncJobChain(NewAsyncJob(nil, db.Disconnect), NewAsyncJobGroup(...))`.
9. Routes live under `/<svc>/`; `/ping`, `/version` are public, `/env` is behind `auth.ApiKeyMiddleware`.
10. Controllers embed `controller.BaseController`; errors map with `errors.Is` to 404/400/500; 5xx never leak error text.
11. `mongo` `Delete` soft-deletes; pass `&repository.QueryOptions{Archived: false}` for a hard delete. Reads hide archived unless `Archived: true`.
12. Tests never call `server.Start()`. Route tests: `chi.NewRouter()` + `r.Group(ms.BuildRoutes)` + `httptest`. Integration tests: `//go:build integration_test`, skip locally, `t.Fatal` when `CI` is set.
13. Mocks are mockery output in `mocks/`; never edit, always regenerate and commit.
14. One `VERSION` file; ldflags set `github.com/bldsoft/gost/version.{Version,GitCommit,GitBranch}`.
15. Lint with the gost `.golangci.yml`; no `//nolint`. Comments only for contracts and invariants with a failure mode.

## Checklists

### Add a configuration variable
- [ ] Field in `pkg/config/config.go` with `mapstructure:"NAME"` and `description:"..."`; secret → `config.HiddenString`.
- [ ] Default in `SetDefaults` (none for secrets); rule in `Validate` naming the variable.
- [ ] Use it in `pkg/microservice` wiring.
- [ ] `make config-md`; commit `config.md`. Add a defaults/validate case to `config_test.go`.

### Add an endpoint
- [ ] Method on the service; extend the interface in `pkg/controller/interface.go`.
- [ ] Handler on the controller: parse (`GetObjectFromBody`, `ParseQueryOption`, `chi.URLParam`) → service → `ResponseJson`/`ResponseOK`; errors via `responseError`.
- [ ] Route in `Controller.Mount` (or in `BuildRoutes` for standard endpoints). Protect admin endpoints.
- [ ] Controller test with a fake service; route test if the path is new.

### Add a repository
- [ ] Entity in `pkg/entity` embedding `mongo.EntityID`, `mongo.EntityTimeStamp`, `mongo.EntityArchived` with `bson:",inline"`.
- [ ] Struct embedding `mongo.Repository[T, *T]` from `mongo.NewRepository[T](db, "coll")`; domain queries as methods.
- [ ] Interface in `pkg/service/interface.go` listing only the methods the service uses; `make generate`.
- [ ] Indexes as a migration (`db.AddMigration(n, up, down)`, JSON command arrays).
- [ ] Integration test behind `integration_test` covering soft-delete and `ErrNotFound`.

### Add a background job
- [ ] `server.NewContextAsyncRunner(func(ctx) error { ticker loop; select on ctx.Done() })`.
- [ ] Construct in `New`, add to the `AsyncJobGroup` in `GetAsyncRunners` (so it stops before the DB disconnects).
- [ ] Unit test with a mock and a short interval; `Run` in a goroutine, `Stop` with a timeout.

### Add a migration
- [ ] Next version number in `pkg/microservice/migrations.go`; both up and down documents.
- [ ] Registered before `Connect`; never edit a shipped migration.

### Add a feature flag
- [ ] `feature.NewFeature(nextStableID, default)` with validator/on-change in the features struct.
- [ ] `feature.AddFeatureMigration(db, version, feature.NewFeature(f, "Name", "Description").WithGroups(...))`.
- [ ] Repository/service/controller wired with the **instance name**; pass the features struct to `GetEnvHandler`.

### Add a mock
- [ ] Interface package listed in `.mockery.yml`; `//go:generate go tool mockery` in `interface.go`.
- [ ] `make generate`; commit `mocks/`. Variadic args are matched as one slice.

## Commands

```bash
make build            # bin/<svc> with version ldflags
make run              # local run with dev defaults
make test             # unit + route tests
make test-integration # needs MongoDB replica set; CI fails, local skips
make lint             # golangci-lint with gost rules
make generate         # mockery, enumer
make config-md        # regenerate config.md
```

## Where to look

| Question | File |
|---|---|
| How is a service started? | [`cmd/notes/main.go`](../examples/notes/cmd/notes/main.go) |
| How is config declared and validated? | [`pkg/config/config.go`](../examples/notes/pkg/config/config.go), [`config_test.go`](../examples/notes/pkg/config/config_test.go) |
| How are repos, services, controllers, runners wired? | [`pkg/microservice/microservice.go`](../examples/notes/pkg/microservice/microservice.go) |
| How are routes tested? | [`pkg/microservice/microservice_test.go`](../examples/notes/pkg/microservice/microservice_test.go) |
| How does a controller parse and respond? | [`pkg/controller/rest/note_controller.go`](../examples/notes/pkg/controller/rest/note_controller.go) |
| How does a repository extend gost? | [`pkg/repository/note_repository.go`](../examples/notes/pkg/repository/note_repository.go) |
| How are integration tests gated? | [`pkg/repository/note_repository_integration_test.go`](../examples/notes/pkg/repository/note_repository_integration_test.go) |
| How does a background job look? | [`pkg/service/purger.go`](../examples/notes/pkg/service/purger.go) |
| What gost packages exist? | [Package reference](package-reference.md) |
| What goes wrong in real services? | [Best practices](best-practices.md) |
