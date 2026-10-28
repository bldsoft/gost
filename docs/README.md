# gost documentation

`github.com/bldsoft/gost` ("GO Service Template") is a toolkit for building Go services: typed environment configuration, a chi based HTTP server with graceful shutdown,
structured logging, MongoDB and ClickHouse storage with migrations, service discovery, runtime
feature flags, auth middlewares, and the small conventions that make all services look alike.
The framework follows the Clean Architecture principles: dependencies point inwards, from HTTP
controllers through services to repositories and domain entities, and each layer consumes
interfaces it declares itself (see [Service anatomy](service-anatomy.md)).

These pages explain how to build a **new service** on gost and how the pieces fit. Every code
snippet is taken from the example service in [`examples/notes`](../examples/notes/README.md), which
the gost CI builds, lints and tests, so what you read here compiles.

## Start here

| You are | Read in this order |
|---|---|
| A developer creating a service | [Quickstart](quickstart.md) → [Service anatomy](service-anatomy.md) → [Configuration](configuration.md) → [Best practices](best-practices.md) |
| A coding agent (Claude Code, Cursor, …) | [Agent guide](agent-guide.md) first, then the page for the area you touch |
| Maintaining an existing service | [Best practices](best-practices.md) → [Package reference](package-reference.md) |

## Guides

| Page | What it covers |
|---|---|
| [Quickstart](quickstart.md) | Create a new service repository step by step, from `go mod init` to green CI |
| [Service anatomy](service-anatomy.md) | Directory layout, layer rules, composition root, startup and shutdown sequence |
| [Configuration](configuration.md) | `config.ReadConfig`, env naming, defaults and validation order, secrets, `config.md` |
| [Server, routing and middleware](server.md) | `server.NewServer`, `IMicroservice`, probes, TLS, default middlewares, known quirks |
| [Logging](logging.md) | `log.InitLogger`, request loggers, `log.FromContext`, exporters |
| [Storage](storage.md) | MongoDB storage, entities, repositories, migrations, soft delete; ClickHouse; caches; locks |
| [Background jobs and shutdown](background-jobs.md) | `server.AsyncRunner`, context runners, job groups and chains, shutdown ordering |
| [Service discovery](discovery.md) | `DISCOVERY_*` config, in-house and Consul backends, discovery aware HTTP clients |
| [Feature flags](feature-flags.md) | Runtime mutable settings with MongoDB persistence and a REST controller |
| [Authentication](auth.md) | API keys, sessions, JWT, IP ACLs |
| [Observability](observability.md) | `/ping`, `/version`, `/env`, `/probes`, `/debug/pprof`, `/stat`, log export, changelog |
| [Testing](testing.md) | Unit tests with mocks or fakes, route tests, integration tests behind a build tag |
| [Build and release](build-release.md) | `VERSION` and ldflags, Dockerfile, Makefile, CI, golangci-lint, mockery, enumer |
| [Best practices](best-practices.md) | Do and don't, with the anti-patterns found in real services and their fixes |
| [Package reference](package-reference.md) | One row per gost package: purpose, entry points, caveats |
| [Agent guide](agent-guide.md) | Invariants and checklists for coding agents |

## Templates

[`docs/templates/`](templates/) holds files to copy into a new service repository: `AGENTS.md`,
`CLAUDE.md`, `Makefile`, `golangci.yml`, `mockery.yml`, `Dockerfile`, `github-workflow.yml`,
`tools.go`, `gitignore`. The [quickstart](quickstart.md) tells you where each one goes.

## Conventions for editing these docs

- Code in the docs must exist in `examples/notes` or in a gost `_test.go` file. Mark the origin
  with an HTML comment such as `<!-- source: examples/notes/pkg/config/config.go -->` so a reader
  can check it.
- When a gost public API changes, update `examples/notes` and the relevant page in the same PR.
- `docs/templates/golangci.yml` must stay byte identical to the root `.golangci.yml`; CI checks it.
- Prefer short pages that answer "how do I" over exhaustive API listings; the
  [package reference](package-reference.md) is the one place for listings.
