# Server, routing and middleware

`server.Server` wraps `net/http` and `chi`. You give it a `server.Config` and one or more
`IMicroservice` values; it mounts routes, adds middlewares, optionally serves Kubernetes probes,
starts your background runners, and shuts everything down on a signal.

## The microservice contract

```go
type Router = chi.Router

type IMicroservice interface {
	BuildRoutes(Router)
	GetAsyncRunners() []AsyncRunner
}

type InitializableMicroservice interface {
	IMicroservice
	Initialize()      // called right before BuildRoutes; connect to databases here
}
```

## Building a server

<!-- source: examples/notes/cmd/notes/main.go -->
```go
server.NewServer(cfg.Server, microservice.New(cfg)).
	UseDefaultMiddlewares().
	WithHealthProbes(true).
	Start()
```

| Builder method | Effect |
|---|---|
| `UseDefaultMiddlewares()` | chi `RequestID`, gost `RealIP`, gost request logger, chi `Recoverer` |
| `AppendMiddlewares(mw...)` | Add global middlewares after the current ones |
| `SetMiddlewares(mw...)` | Replace the global middleware chain |
| `SetRouterWrapper(fn)` | Wrap the whole application router (for CORS or per-prefix middleware chains) |
| `WithHealthProbes(true, internal...)` | Mount `/probes/healthy`, `/probes/ready`, `/probes/startup`; with an `InternalServer` or `PROBES_BIND_ADDRESS` they move to the private port (see [Probes on a private port](#probes-on-a-private-port)) |
| `AddAsyncRunners(r...)` | Register runners created in `main` (in addition to `GetAsyncRunners`) |
| `Start()` | Listen and block until SIGINT/SIGTERM/SIGQUIT |

Always mounted: `/debug/pprof/*` (chi profiler plus `goroutineleak`). Protect it at the ingress.

## Routing conventions

Each microservice gets its own `chi` group, so `r.Use(...)` inside `BuildRoutes` applies to that
microservice only. Mount everything under a prefix equal to the service name:

<!-- source: examples/notes/pkg/microservice/microservice.go -->
```go
r.Route(BasePath, func(r server.Router) {             // "/notes"
	r.Get("/ping", gost.GetPingHandler)
	r.Get("/version", gost.GetVersionHandler)
	r.With(auth.ApiKeyMiddleware(APIKeyHeader, m.cfg.AdminAPIKey.String())).
		Get("/env", gost.GetEnvHandler(&m.cfg, nil))

	r.Route("/api/v1/notes", m.notes.Mount)
})
```

Controllers expose `Mount(r chi.Router)` so the composition root decides the path:

<!-- source: examples/notes/pkg/controller/rest/note_controller.go -->
```go
func (c *NoteController) Mount(r chi.Router) {
	r.Get("/", c.List)
	r.Post("/", c.Create)
	r.Get("/{id}", c.Get)
	r.Put("/{id}", c.Update)
	r.Delete("/{id}", c.Delete)
}
```

## Writing handlers

Embed `controller.BaseController` (import gost's `controller` package under an alias such as
`gost` when your own package is also called `controller`):

| Helper | Use |
|---|---|
| `c.GetObjectFromBody(w, r, &obj) bool` | Decode JSON body; writes 400 and returns false on error |
| `c.ResponseJson(w, r, v)` | Marshal `v`, set `Content-Type: application/json` |
| `c.ResponseError(w, msg, code)` | `http.Error`; for 5xx also attaches the message to the exported request log |
| `c.ResponseOK(w)` | Write `OK` |
| `controller.ParseQueryOption(r, w, "limit", &limit) bool` | Typed query param with default; writes 400 on parse error |
| `controller.GetQueryOption[T](r, name, default)` | Same, returning the error to you |
| `controller.GetQueryOptionSlice[T]`, `ParseQueryOptionSlice` | Comma separated lists |
| `utils.FromRequest[F](r)` | Decode the whole query string into a struct with gorilla/schema |
| `controller.QueryOptionFromRequest[F](r, archived)` | Build `repository.QueryOptions` with a typed filter and `?fields=` |
| `chi.URLParam(r, "id")` | Path parameters |

Map errors to status codes in one place per controller:

<!-- source: examples/notes/pkg/controller/rest/note_controller.go -->
```go
func (c *NoteController) responseError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		c.ResponseError(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	case errors.Is(err, entity.ErrInvalidNote):
		c.ResponseError(w, err.Error(), http.StatusBadRequest)
	default:
		log.FromContext(r.Context()).ErrorWithFields(log.Fields{"error": err}, "Note request failed")
		c.ResponseError(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
```

Never leak internal error text on 5xx; log it with the request logger instead.

## Standard endpoints

| Path | Handler | Auth |
|---|---|---|
| `/<svc>/ping` | `controller.GetPingHandler` → `pong` | none |
| `/<svc>/version` | `controller.GetVersionHandler` → `{"version","branch","commit"}` | none |
| `/<svc>/env` | `controller.GetEnvHandler(&cfg, features)` → YAML of the config (+ features) | API key, always |
| `/probes/{healthy,ready,startup}` | built in; on `PROBES_BIND_ADDRESS` when set | none |
| `/debug/pprof/*` | built in | protect at ingress |

`ready` returns 503 until the application router (and therefore `Initialize()`) has finished.

## Probes on a private port

The ingress should never see the probes. Set `PROBES_BIND_ADDRESS` (for example `0.0.0.0:3002`)
and `WithHealthProbes(true)` serves `/probes/*` only there; the main listener answers 404 on
those paths. Readiness keeps its meaning: the internal listener binds in `Start()` together with
the main one while the application router is still being built, `/probes/ready` is 503 until
`Initialize()` and `BuildRoutes` finish, then 200. The listener is shut down with the server
inside the 5 second budget.

Pass your own `server.InternalServer` when the private port should also carry other
platform-facing handlers such as Prometheus metrics. Probes are enabled in exactly one place,
`WithHealthProbes`, because only the `Server` knows when the service is ready:

<!-- source: server/internal_server_test.go -->
```go
internal := NewInternalServer(cfg.ProbesBindAddress).   // server.NewInternalServer(cfg.Server.ProbesBindAddress) in main.go
	Handle("/metrics", metrics)                           // for example promhttp.Handler()
srv := NewServer(cfg, ms).
	UseDefaultMiddlewares().
	WithHealthProbes(true, internal)
```

Rules of the internal listener:

- It has its own `chi` router (`Handle`, `Mount`) and never touches `http.DefaultServeMux`.
- The common middlewares do not apply, so probe hits do not show up in the request log.
- Timeouts are fixed (5 s header, 10 s read, 30 s write, 60 s idle) regardless of
  `READ_TIMEOUT_MS`/`WRITE_TIMEOUT_MS`, which are often `0` for long-running app requests.
- An `InternalServer` with an empty address is disabled with a warning and the probes stay on
  the main listener, so `NewInternalServer(cfg.Server.ProbesBindAddress)` is safe when the
  variable is unset.
- `InternalServer` is also an `AsyncRunner`: extra probe-less listeners can be registered with
  `AddAsyncRunners(internal)`. They start after the routes are built, like any runner.

## Middleware from gost

| Middleware | Package | Notes |
|---|---|---|
| `RealIP` | `server/middleware` | chi RealIP plus `GetRealIP(ctx)` / `WithRealIP(ctx, ip)` |
| `RequestID` | `server/middleware` | also accepts `?req-id=` |
| Request logger | `log` | `log.DefaultRequestLogger()`, `NewRequestLogger(formatter)`, `MultiLogger(...)`; skip paths with `log.NotLoggedEndpoints` |
| `auth.ApiKeyMiddleware(header, key)` | `auth` | 401 on mismatch |
| `jwt.JwtAuthMiddlewareFromConfig(cfg)` | `auth/jwt` | |
| `acl.MiddlewareFromConfig(cfg)` | `auth/acl` | IP allow/deny lists |

## TLS and timeouts

Setting `TLS_CERTIFICATE_PATH` and `TLS_KEY_PATH` adds an HTTPS listener on
`TLS_SERVICE_BIND_ADDRESS` (default `<host>:3443`). `READ_TIMEOUT_MS` and `WRITE_TIMEOUT_MS` map
to `http.Server` timeouts; `0` disables them.

## Known quirks

These are properties of the current `server` package. Design around them; do not work around
them in tests.

- **Global mux.** `Start()` registers the handler on `http.DefaultServeMux`. One `server.Server`
  per process, and tests must build a `chi.NewRouter()` and call `BuildRoutes` instead of
  starting the server. Anything else registered on `DefaultServeMux` becomes reachable.
- **Shutdown context.** `http.Server.Shutdown` is called with the already cancelled signal
  context, so in-flight requests are not drained. Runners do get a 5 second budget.
- **Listener errors do not abort.** A failed `net.Listen` is logged and the process keeps
  running until a signal arrives. Watch the log line `Failed to create http listener`.
- **Runners start after routes.** `GetAsyncRunners()` is collected after `BuildRoutes`. A
  runner that must be running before traffic arrives has to start in `Initialize()`.
- **Internal listener errors do not abort either.** When `PROBES_BIND_ADDRESS` cannot be bound,
  the probes are served nowhere: the main router already skipped them. Watch the log line
  `Failed to create internal listener`.
