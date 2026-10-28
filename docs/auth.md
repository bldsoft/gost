# Authentication and authorisation

gost offers four mechanisms. Pick the simplest that fits; most internal services need only an
API key for admin endpoints.

| Need | Use |
|---|---|
| Protect admin endpoints (`/env`, dumps, maintenance) | `auth.ApiKeyMiddleware(header, key)` |
| Browser or operator login with sessions | `auth.NewAuthController` + a session store |
| Service to service tokens | `auth/jwt` |
| Restrict by caller IP | `auth/acl` |

## API key

<!-- source: examples/notes/pkg/microservice/microservice.go -->
```go
r.With(auth.ApiKeyMiddleware(APIKeyHeader, m.cfg.AdminAPIKey.String())).
	Get("/env", gost.GetEnvHandler(&m.cfg, nil))
```

The middleware compares the header value with the configured key and answers 401 otherwise.
Conventions: header named `<SVC>-API-Key` or `X-API-Key`; the key is a `config.HiddenString`
with no default; the service refuses to start without one (see
[Configuration](configuration.md#secrets)). Letting loopback peers through without a key is a
reasonable local-debug convenience.

## Sessions

Generic over your user type. The user type embeds `auth.User` (login + bcrypt password) and
`mongo.EntityID`:

```go
type Account struct {
	mongo.EntityID `bson:",inline"`
	auth.User      `bson:",inline"`
}

store, _ := auth_store.NewMongoDBStore(db, []byte(cfg.SessionKey.String()))    // or NewRedisStore(client)
authService := auth.NewAuthService[*Account](accountRepo, &auth.BcryptHasher{})
authController := auth.NewAuthController[*Account](authService, store, "sid")

r.Route("/auth", authController.Mount)                        // POST /login, /logout, /refresh
r.Group(func(r chi.Router) {
	r.Use(authController.AuthenticateMiddleware())            // puts the user into the context; does NOT reject anonymous requests
	r.Use(requireUser)                                        // your check on auth.UserFromContext(ctx)
	...
})
```

`auth.NewUserService` and `auth.NewUserController` add user CRUD (`GET /`, `GET /{id}`,
`POST /`, `PUT /{id}`, `PUT /{id}/password`, `DELETE /{id}`). Sentinel errors
`auth.ErrUnauthorized` and `auth.ErrForbidden` are meant to be mapped to 401 and 403 in your
controllers.

## JWT

```go
type Config struct {
	Jwt jwt.JwtConfig `mapstructure:"JWT"`      // JWT_ALG, JWT_KEY_PATH (PEM); Validate loads the key
}

r.Use(jwt.JwtAuthMiddlewareFromConfig(cfg.Jwt))
signed, err := jwt.SignToken(token, cfg.Jwt)
```

## IP ACL

```go
type Config struct {
	Acl acl.Config `mapstructure:"ACL"`         // ACL_ALLOW, ACL_DENY: comma separated IPs or CIDRs
}

r.Use(acl.MiddlewareFromConfig(cfg.Acl))        // uses the real IP from server/middleware.RealIP
```

`acl.IpRange` is a fast prefix set (bart tree) with JSON and BSON marshalling, reusable for
your own allow lists.
