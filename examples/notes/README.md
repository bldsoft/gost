# notes: the gost example service

A deliberately small service that shows the canonical way to build on
[gost](../../README.md). It stores notes in MongoDB, soft-deletes them, and purges old
soft-deleted notes in the background. Every page in [`docs/`](../../docs/README.md) links to code
in this directory, and the gost CI builds, lints and tests it, so the snippets cannot rot.

What it demonstrates:

| Concern | Where |
|---|---|
| Entry point: read config, init logger, start the server | [`cmd/notes/main.go`](cmd/notes/main.go) |
| Config struct composed from gost configs, defaults, validation, a masked secret | [`pkg/config/config.go`](pkg/config/config.go) |
| `config.md` kept in sync by a golden test | [`pkg/config/config_test.go`](pkg/config/config_test.go), [`config.md`](config.md) |
| Entity with gost base types (ID, timestamps, soft delete) | [`pkg/entity/note.go`](pkg/entity/note.go) |
| Mongo repository on top of `mongo.Repository`, hard delete for the purge | [`pkg/repository/note_repository.go`](pkg/repository/note_repository.go) |
| Service that declares the repository interface it needs | [`pkg/service/interface.go`](pkg/service/interface.go), [`pkg/service/note_service.go`](pkg/service/note_service.go) |
| Background job as a `server.AsyncRunner` | [`pkg/service/purger.go`](pkg/service/purger.go) |
| REST controller on `controller.BaseController`, error to status mapping | [`pkg/controller/rest/note_controller.go`](pkg/controller/rest/note_controller.go) |
| Composition root: wiring, migrations, routes, shutdown order | [`pkg/microservice/microservice.go`](pkg/microservice/microservice.go) |
| Tests: mockery mocks, hand-written fakes, route tests, integration tests | `*_test.go` next to the code |

## Run it

```bash
# MongoDB must run as a replica set (change streams and transactions need one).
docker run -d --name mongo -p 27017:27017 mongo:8 --replSet rs0 --bind_ip_all
docker exec mongo mongosh --quiet --eval 'rs.initiate({_id:"rs0",members:[{_id:0,host:"127.0.0.1:27017"}]})'

make run            # ADMIN_API_KEY=dev-key, DATABASE_SERVER and DATABASE_DBNAME have local defaults
```

In another terminal:

```bash
curl -s localhost:3000/notes/ping                          # pong
curl -s localhost:3000/notes/version                       # {"version":"0.1.0",...}
curl -s -o /dev/null -w '%{http_code}\n' localhost:3000/notes/env   # 401
curl -s -H 'X-API-Key: dev-key' localhost:3000/notes/env   # YAML config dump, AdminAPIKey: '****'
curl -s localhost:3000/probes/ready                        # OK

id=$(curl -s -X POST localhost:3000/notes/api/v1/notes -d '{"title":"first","body":"hello"}' | jq -r .id)
curl -s localhost:3000/notes/api/v1/notes/$id
curl -s -X PUT localhost:3000/notes/api/v1/notes/$id -d '{"title":"renamed"}'
curl -s localhost:3000/notes/api/v1/notes?limit=10
curl -s -X DELETE localhost:3000/notes/api/v1/notes/$id    # OK (soft delete)
curl -s -o /dev/null -w '%{http_code}\n' localhost:3000/notes/api/v1/notes/$id   # 404
curl -s "localhost:3000/notes/api/v1/notes/$id?archived=true"                    # still there, with deleteTime
```

Stop it with Ctrl-C: the purger and discovery stop first, then MongoDB disconnects, then the
process exits with "Server gracefully stopped".

## Make targets

```text
make test              unit and route tests, no database needed
make test-integration  needs MongoDB; skips locally when unreachable, fails in CI
make lint              golangci-lint with the gost configuration
make generate          regenerate mocks (go tool mockery from the gost root)
make config-md         regenerate config.md from the Config struct
make docker            docker build with the gost repository root as context
```

## Differences from a standalone service

The example is a package inside the gost module, so a few things differ from the repo you create
with the [quickstart](../../docs/quickstart.md):

- Import paths start with `github.com/bldsoft/gost/examples/notes/`; yours start with your module.
- There is no `go.mod` here; yours pins `github.com/bldsoft/gost` at a version.
- The Dockerfile uses the gost root as the build context; a service from `scaffold` uses
  [`examples/scaffold/Dockerfile`](../scaffold/Dockerfile), which builds from its own repository root.
- Mocks are configured in the gost `.mockery.yml`; yours has its own
  ([`examples/scaffold/mockery.yml`](../scaffold/mockery.yml)).
