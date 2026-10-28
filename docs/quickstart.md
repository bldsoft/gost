# Quickstart: a new service on gost

This walkthrough creates a service called `<svc>` in a module `github.com/<org>/<svc>` that
serves `/<svc>/ping`, `/<svc>/version`, `/<svc>/env`, Kubernetes probes, and one domain resource
stored in MongoDB. It mirrors [`examples/notes`](../examples/notes/README.md) exactly, so when a
step says "copy from the example", open that file and rename.

Placeholders used below:

| Placeholder | Meaning | Example |
|---|---|---|
| `<org>` | GitHub organisation or user | `acme` |
| `<svc>` | Service name, lower case, also the binary name and URL prefix | `widgets` |

## 0. Prerequisites

- Go 1.27 or newer (gost's `go.mod` sets the floor).
- Docker, for a local MongoDB replica set and for image builds.
- `golangci-lint` v2.15 or newer built with the same Go major version as your toolchain.
- Optional: `gotestsum` for nicer test output.

## 1. Create the module

```bash
mkdir <svc> && cd <svc> && git init
go mod init github.com/<org>/<svc>
go get github.com/bldsoft/gost@latest               # or a specific commit/tag; pin ONE version per service
go get github.com/hashicorp/go-metrics@v0.5.4       # see note below
echo "0.1.0" > VERSION
```

The second `go get` is needed because `go mod tidy` in a module that imports gost's discovery
packages resolves the tests of `hashicorp/consul/api`, which reach `hashicorp/memberlist` and
`hashicorp/go-metrics/compat`; gost itself does not need that module, so it cannot pin it for
you. Without the line, `go mod tidy` fails with
`module github.com/hashicorp/go-metrics@latest found (v0.7.0), but does not contain package .../compat`.

## 2. Copy the templates

`GOST` is a checkout of this repository.

```bash
GOST=/path/to/gost
mkdir -p .github/workflows tools cmd/<svc> pkg/{config,entity,repository,service,controller/rest,microservice}
cp $GOST/docs/templates/Makefile            Makefile
cp $GOST/docs/templates/golangci.yml        .golangci.yml
cp $GOST/docs/templates/mockery.yml         .mockery.yml
cp $GOST/docs/templates/Dockerfile          Dockerfile
cp $GOST/docs/templates/github-workflow.yml .github/workflows/ci.yml
cp $GOST/docs/templates/gitignore           .gitignore
cp $GOST/docs/templates/tools.go            tools/tools.go
cp $GOST/docs/templates/AGENTS.md           AGENTS.md
cp $GOST/docs/templates/CLAUDE.md           CLAUDE.md
```

Replace the placeholders in every copied file (`grep -rn '<org>\|<svc>' .` finds them). If your
module path does not share gost's prefix, change the `prefix(...)` entry in the `gci` section of
`.golangci.yml` to your own module path so imports group correctly. Then pin the code
generators as Go tools so `go tool mockery` and `go tool enumer` work without a global install:

```bash
go get -tool github.com/vektra/mockery/v3@latest
go get -tool github.com/dmarkham/enumer@latest
```

## 3. Write the four core files

Copy these from the example and rename `notes` to `<svc>`:

| Copy from | To | Change |
|---|---|---|
| [`cmd/notes/main.go`](../examples/notes/cmd/notes/main.go) | `cmd/<svc>/main.go` | import paths |
| [`pkg/config/config.go`](../examples/notes/pkg/config/config.go) | `pkg/config/config.go` | `ServiceName`, drop `NOTES_PURGE_*` if you have no purger |
| [`pkg/config/config_test.go`](../examples/notes/pkg/config/config_test.go) | `pkg/config/config_test.go` | assertions on your defaults |
| [`pkg/microservice/microservice.go`](../examples/notes/pkg/microservice/microservice.go) | `pkg/microservice/microservice.go` | wire your own repositories, services and controllers |

The entry point is always the same five lines:

<!-- source: examples/notes/cmd/notes/main.go -->
```go
var cfg config.Config
gost.ReadConfig(&cfg, "")                                      // defaults → env → validate; fatal on error
log.InitLogger(&cfg.Log, cfg.Server.LogExporterConfig())

log.Infof("%s v%s", config.ServiceName, version.Version)
log.Logf("ENVIRONMENT:\n***\n%s***", gost.FormatEnv(&cfg))     // secrets are masked by HiddenString

server.NewServer(cfg.Server, microservice.New(cfg)).
	UseDefaultMiddlewares().
	WithHealthProbes(true).
	Start()
```

The config struct embeds gost sub-configs and adds your own fields:

<!-- source: examples/notes/pkg/config/config.go -->
```go
type Config struct {
	Server    server.Config
	Log       log.Config
	Mongo     mongo.Config
	Discovery common.Config `mapstructure:"DISCOVERY"`

	AdminAPIKey config.HiddenString `mapstructure:"ADMIN_API_KEY" description:"API key expected in the X-API-Key header by /notes/env. Required: the service refuses to start without it."`
}

func (c *Config) SetDefaults() {
	c.Server.ServiceName = ServiceName          // mandatory: server.Config.Validate fails without it
	c.Server.ServiceBindAddress = "0.0.0.0:3000"
	c.Mongo.DbName = ServiceName
}

func (c *Config) Validate() error {
	if c.AdminAPIKey == "" {
		return errors.New("ADMIN_API_KEY is required")   // empty secret fails closed
	}

	return nil
}
```

Do not call `c.Server.SetDefaults()` or `c.Mongo.Validate()` yourself: `ReadConfig` walks the
whole struct and calls them for you, children first. See [Configuration](configuration.md).

## 4. Start MongoDB

gost's Mongo features (change streams, transactions, changelog) need a replica set, so run one
even locally:

```bash
docker run -d --name mongo -p 27017:27017 mongo:8 --replSet rs0 --bind_ip_all
docker exec mongo mongosh --quiet --eval 'rs.initiate({_id:"rs0",members:[{_id:0,host:"127.0.0.1:27017"}]})'
```

## 5. Run and poke it

```bash
make run     # the Makefile exports ADMIN_API_KEY=dev-key and a local DATABASE_SERVER
```

```bash
curl -s localhost:3000/<svc>/ping                               # pong
curl -s localhost:3000/<svc>/version                            # {"version":"0.1.0", ...}
curl -s -o /dev/null -w '%{http_code}\n' localhost:3000/<svc>/env   # 401 without the key
curl -s -H 'X-API-Key: dev-key' localhost:3000/<svc>/env        # YAML dump, AdminAPIKey: '****'
curl -s localhost:3000/probes/ready                             # OK once the DB is connected
```

Ctrl-C stops it: runners stop, MongoDB disconnects, "Server gracefully stopped" is logged.

## 6. Add your first resource

Follow the layer order; each step has a one-file example:

1. **Entity**: a struct embedding `mongo.EntityID`, `mongo.EntityTimeStamp`, `mongo.EntityArchived`
   ([`pkg/entity/note.go`](../examples/notes/pkg/entity/note.go)).
2. **Repository**: embed `mongo.Repository[T, *T]` created with `mongo.NewRepository[T](db, "collection")`
   ([`pkg/repository/note_repository.go`](../examples/notes/pkg/repository/note_repository.go)).
3. **Service**: declare the repository interface you consume in `pkg/service/interface.go`, add
   `//go:generate go tool mockery`, implement the use cases
   ([`pkg/service/note_service.go`](../examples/notes/pkg/service/note_service.go)).
4. **Controller**: declare the service interface in `pkg/controller/interface.go`; the REST
   controller embeds `controller.BaseController` and maps errors to status codes with `errors.Is`
   ([`pkg/controller/rest/note_controller.go`](../examples/notes/pkg/controller/rest/note_controller.go)).
5. **Wire** it in `pkg/microservice/microservice.go`: construct in `New`, mount under
   `/<svc>/api/v1/...` in `BuildRoutes`, register migrations with `db.AddMigration` before connect.
6. **Tests**: `make generate` for mocks (before `go mod tidy`, since tests import `mocks/`),
   then unit, route and integration tests as in [Testing](testing.md).

## 7. Keep `config.md` in sync

Every env var is documented from the `description` tags. The golden test in
`pkg/config/config_test.go` fails when `config.md` is stale:

```bash
make config-md && git add config.md
```

## 8. Lint, test, push

```bash
make lint && make test && make test-integration
git add -A && git commit -m "feat: bootstrap <svc> on gost"
```

The CI template runs lint, unit and integration tests (with a MongoDB replica set), checks that
mocks are up to date, and builds the image on `main` and `v*` tags. A tag must match `VERSION`.

## Troubleshooting

| Symptom | Cause and fix |
|---|---|
| `Invalid config: ServiceName is not set. Do it in SetDefaults method` | Set `c.Server.ServiceName` in your `Config.SetDefaults` |
| `Invalid config: ADMIN_API_KEY is required` | Export the variable; an empty secret is rejected on purpose |
| `error connecting to db` repeating every second | MongoDB is down or `DATABASE_SERVER` is wrong; `DBConnectAsync` retries forever and `/probes/ready` stays 503 until it succeeds |
| `Failed to create http listener ... address already in use` | Change `SERVICE_BIND_ADDRESS`; note the process keeps running and shuts down on signal |
| `SERVICE_NAME is deprecated, use SERVICE_INSTANCE_NAME instead` | Harmless warning; set `SERVICE_INSTANCE_NAME` per instance in deployment |
| Integration tests are skipped | MongoDB unreachable locally; set `DATABASE_SERVER`. In CI (`CI=true`) they fail instead |
| `go mod tidy`: `go-metrics@latest ... does not contain package .../compat` | Run `go get github.com/hashicorp/go-metrics@v0.5.4` (step 1) |
| `go mod tidy`: `repository 'https://github.com/<org>/<svc>/' not found` | A test imports `pkg/service/mocks` that does not exist yet; run `go tool mockery` first, then tidy |
