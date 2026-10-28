# Build, lint and release

Everything below is in the [templates](templates/) and used by [`examples/notes`](../examples/notes).

## Version

One `VERSION` file at the repository root (`0.1.0`), bumped by hand in the release commit. The
binary learns it through ldflags on the gost `version` package:

<!-- source: examples/notes/Makefile -->
```make
VERSION    := $(shell cat VERSION)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)
LDFLAGS    := -X github.com/bldsoft/gost/version.Version=$(VERSION) \
              -X github.com/bldsoft/gost/version.GitCommit=$(GIT_COMMIT) \
              -X github.com/bldsoft/gost/version.GitBranch=$(GIT_BRANCH)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/notes ./cmd/notes
```

Some services also `//go:embed VERSION` and assign `version.Version` in `main`; that works but
keep a single `VERSION` file either way (two copies inevitably disagree). Tags are
`v<VERSION>`; the CI template refuses a tag that does not match the file.

## Makefile

Targets every service should have (see [`docs/templates/Makefile`](templates/Makefile)):

| Target | Does |
|---|---|
| `build` | `go build` with ldflags into `bin/` |
| `run` | build and run with local defaults exported |
| `test` | `go test -race -count=1 ./...` |
| `test-integration` | same with `-tags=integration_test` |
| `lint` | `golangci-lint run ./...` |
| `generate` | `go generate ./...` (mockery, enumer) |
| `config-md` | regenerate `config.md` through the golden test |
| `docker` | build the image with commit and branch build args |

## Dockerfile

Multi-stage, static binary, non-root, version injected ([`docs/templates/Dockerfile`](templates/Dockerfile)):

```dockerfile
FROM golang:1.27-alpine AS builder
ARG COMMIT=unknown
ARG BRANCH=unknown
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w \
      -X github.com/bldsoft/gost/version.Version=$(cat VERSION) \
      -X github.com/bldsoft/gost/version.GitCommit=${COMMIT} \
      -X github.com/bldsoft/gost/version.GitBranch=${BRANCH}" \
    -o /out/<svc> ./cmd/<svc>

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/<svc> /usr/local/bin/<svc>
USER 65532:65532
EXPOSE 3000
ENTRYPOINT ["/usr/local/bin/<svc>"]
```

Add a `.dockerignore` with `.env*`, `bin/`, `.git/` so secrets and build output never enter a
layer. The example's own Dockerfile differs only in using the gost repository root as context.

## CI

[`docs/templates/github-workflow.yml`](templates/github-workflow.yml) has three jobs:

1. **lint**: `golangci/golangci-lint-action@v9` pinned to `v2.15.0`, using `go-version-file: go.mod`.
2. **test**: starts `mongo:8` as a single-node replica set (ping, `rs.initiate`, wait for
   primary), installs gotestsum, runs `gotestsum -- -race -tags=integration_test ./...`, then
   `go tool mockery && git diff --exit-code` to catch stale mocks.
3. **docker**: on pushes to `main` and `v*` tags, checks the tag matches `VERSION` and builds
   the image with `COMMIT` and `BRANCH` build args. Add your registry login and push.

Jenkins based services do the same with `--build-arg BUILD/BRANCH/COMMIT`; the only thing that
matters is that the ldflags are set.

## golangci-lint

Copy [`docs/templates/golangci.yml`](templates/golangci.yml) to `.golangci.yml` unchanged; it is
the gost configuration and CI keeps the two identical. Rules that bite most:

| Linter | Rule | Fix |
|---|---|---|
| `nlreturn` | a `return` preceded by other statements in the same block needs a blank line before it | add the blank line |
| `govet shadow` (strict) | no shadowing except `err`, `ctx`, `ok` | rename the inner variable |
| `gci` | imports in three groups: standard, third party, the configured module `prefix(...)` | `golangci-lint fmt` or `gci write` |
| `errorlint` | compare errors with `errors.Is` / `errors.As` | never `err == ErrX` |
| `gosmopolitan` | no Cyrillic in string literals | English only |
| `errcheck` | check every error, including `Close`, `Write` | `_ = f.Close()` when ignoring is intended |

Generated code in `mocks/` and `mock/` is excluded. Never add `//nolint` without a reviewer
asking for it. Build the linter with the same Go major version as your toolchain, or
`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0 run ./...`.

## Code generation

| Tool | Directive | Output |
|---|---|---|
| mockery v3 | `//go:generate go tool mockery` above an interface file; packages listed in `.mockery.yml` | `<pkg>/mocks/<Interface>.go`, testify template with expecters |
| enumer | `//go:generate go tool enumer -type=Status -json -text` | `status_enumer.go` with `String`, `ParseStatus`, marshalling |

Pin both as Go tools (`go get -tool ...`), which adds `tool` lines to `go.mod`, and keep
`tools/tools.go` (build tag `tools`) so `go mod tidy` does not drop packages that production code
does not import yet. Mockery discovers `.mockery.yml` by walking up from the current directory,
so `go tool mockery` works from any package.

## Pull requests

gost's PR template asks for a summary, the ticket key, the type of change, a checklist that
includes "I've manually reviewed any AI generated code", and a GIF. Copy it to your service if
you want the same discipline. Before pushing: `make lint && make test && make test-integration`.
