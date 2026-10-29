# gost

**GO Service Template**: a toolkit for building Go services. It gives a new service
typed environment configuration, a chi based HTTP server with graceful shutdown, structured
logging, MongoDB and ClickHouse storage with migrations, service discovery, runtime feature flags,
auth middlewares and the shared conventions (`/ping`, `/version`, `/env`, probes, `config.md`,
`VERSION`) that make every service built on it look the same to operators.

Services built on gost follow the Clean Architecture principles: domain entities at the centre,
business logic in services that depend only on interfaces they declare, storage and HTTP as
replaceable outer layers, and a single composition root that wires them together. See
[Service anatomy](docs/service-anatomy.md) for the layer rules.

## Documentation

- **[Create a new service](docs/quickstart.md)**: step by step, from `go mod init` to green CI.
- **[Documentation index](docs/README.md)**: anatomy, configuration, server, logging, storage,
  background jobs, discovery, feature flags, auth, observability, testing, build and release,
  best practices, package reference.
- **[Agent guide](docs/agent-guide.md)**: invariants and checklists for coding agents.
- **[Example service](examples/notes/README.md)**: a complete, tested service in `examples/notes`
  that every doc page links to. CI builds, lints and tests it.
- **[Templates](docs/templates/)**: `AGENTS.md`, `Makefile`, `.golangci.yml`, `Dockerfile`, CI
  workflow and more to copy into a new repository.

## A service in 25 lines

<!-- source: examples/notes/cmd/notes/main.go, examples/notes/pkg/config/config.go -->
```go
type Config struct {
	Server server.Config
	Log    log.Config
	Mongo  mongo.Config

	AdminAPIKey config.HiddenString `mapstructure:"ADMIN_API_KEY" description:"API key for /notes/env"`
}

func (c *Config) SetDefaults() {
	c.Server.ServiceName = "notes"
	c.Server.ServiceBindAddress = "0.0.0.0:3000"
	c.Mongo.DbName = "notes"
}

func (c *Config) Validate() error {
	if c.AdminAPIKey == "" {
		return errors.New("ADMIN_API_KEY is required")
	}

	return nil
}

func main() {
	var cfg Config
	config.ReadConfig(&cfg, "")
	log.InitLogger(&cfg.Log, cfg.Server.LogExporterConfig())

	server.NewServer(cfg.Server, microservice.New(cfg)).
		UseDefaultMiddlewares().
		WithHealthProbes(true).
		Start()
}
```

`microservice.New(cfg)` returns a `server.IMicroservice` that wires repositories, services and
controllers and mounts routes; see [`examples/notes/pkg/microservice`](examples/notes/pkg/microservice/microservice.go).

## Packages

| Area | Packages |
|---|---|
| Configuration | `config`, `config/feature` |
| HTTP | `server`, `server/middleware`, `controller`, `auth`, `auth/jwt`, `auth/acl`, `auth/store` |
| Logging | `log` |
| Storage | `mongo`, `mongo-lock`, `distlock`, `clickhouse`, `storage`, `repository`, `cache`, `cache/v2` |
| Platform | `discovery`, `feature`, `changelog`, `stat`, `alert`, `breaker`, `version` |
| Utilities | `utils/*` (errgroup, exporter, worker_pool, poly, health_check, ...) |

The [package reference](docs/package-reference.md) lists entry points and caveats for each.

## Requirements

Go 1.27 or newer (see `go.mod`). Integration tests need MongoDB 8 as a replica set, memcached
and Consul; `.github/workflows/gost.yml` shows how CI starts them.

## Development

```bash
make lint        # golangci-lint run ./...   (v2.15+, built with the same Go major version)
make test        # gotestsum -race -tags=integration_test ./...   (needs the services above)
make generate    # go generate ./...  (mockery via `go tool mockery`, enumer)
make example     # go build ./examples/...
```

## License

MIT, see [LICENSE](LICENSE).

## Contributing

Pull requests follow `.github/PULL_REQUEST_TEMPLATE.md`. When you change a public API, update
`examples/notes` and the relevant page in `docs/` in the same PR; the example is part of the root
module, so `go build ./...` catches the drift.
