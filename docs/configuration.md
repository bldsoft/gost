# Configuration

gost services are configured **only through environment variables**. The `config` package reads
a nested struct with viper, applies defaults, validates, and can print a Markdown table of every
variable. The example lives in [`examples/notes/pkg/config`](../examples/notes/pkg/config/config.go).

## The contract

```go
type IConfig interface {
	SetDefaults()
	Validate() error
}
```

`config.ReadConfig(&cfg, envPrefix)` does four things, in this order, and calls `log.Fatalf` on
any failure:

1. `SetDefaults()` on every nested `IConfig`, **children before parents**. A parent can therefore
   override a child's default (for example `c.Server.ServiceName`, `c.Mongo.DbName`).
2. If `CONFIG_DESCRIPTION_PATH` is set, write the Markdown description of all variables there.
3. Read the environment into the struct.
4. `Validate()` on every nested `IConfig`, children first.

Because gost walks the whole tree, **never call nested `SetDefaults()` or `Validate()` yourself**.
Doing so is harmless but misleading, and it is a common anti-pattern in older services.

## Environment variable names

The name of a leaf field is its `mapstructure` tag, or the upper-cased field name if there is no
tag. The prefix is `envPrefix` plus the `mapstructure` tags of all parent structs, joined with
`_`. An untagged or embedded struct adds no prefix segment.

<!-- source: config/example_test.go -->
```go
type Config struct {
	Param1     string                                   // EX_PARAM1
	Tagged     string     `mapstructure:"TAGGED"`       // EX_TAGGED
	Mongo      Database   `mapstructure:"MONGO"`        // EX_MONGO_CONNECTION_STRING
	Clickhouse Clickhouse `mapstructure:"CLICKHOUSE"`   // EX_CLICKHOUSE_CONNECTION_STRING, EX_CLICKHOUSE_PARAM1
}
// ReadConfig(&cfg, "EX")
```

Conventions used by all services:

| Field | Tag | Resulting variables |
|---|---|---|
| `Server server.Config` | none | `SERVICE_INSTANCE_NAME`, `SERVICE_BIND_ADDRESS`, `SERVICE_ADDRESS`, `TLS_*`, `READ_TIMEOUT_MS`, `WRITE_TIMEOUT_MS` |
| `Log log.Config` | none | `LOG_LEVEL`, `LOG_COLOR_ENABLED`, `LOG_FILE` |
| `Mongo mongo.Config` | none | `DATABASE_SERVER`, `DATABASE_DBNAME`, `DATABASE_MIGRATION_COLLECTION` |
| `Clickhouse clickhouse.Config` | `mapstructure:"CLICKHOUSE"` | `CLICKHOUSE_DSN` (the struct's own tag is just `DSN`) |
| `LogExport clickhouse.LogExporterConfig` | none | `CLICKHOUSE_LOG_EXPORT_*` |
| `Discovery common.Config` | `mapstructure:"DISCOVERY"` | `DISCOVERY_TYPE`, `DISCOVERY_INHOUSE_*`, `DISCOVERY_CONSUL_*` |

Most services pass an empty `envPrefix`. A prefix such as `"SP"` makes the variables read
`SP_SERVICE_BIND_ADDRESS`; pick one style per service and never mix.

Supported field types: strings and named string types (`config.Address`, `config.HiddenString`,
`config.ConnectionString`), integers, booleans, `time.Duration` (Go syntax such as `1h30m`),
and slices from comma separated values (`a,b,c`).

## Mandatory defaults

`server.Config.ServiceName` has no env tag (`description:"-"`) and `server.Config.Validate`
rejects an empty value, so every service sets it in its own `SetDefaults`:

<!-- source: examples/notes/pkg/config/config.go -->
```go
func (c *Config) SetDefaults() {
	c.Server.ServiceName = ServiceName
	c.Server.ServiceBindAddress = "0.0.0.0:3000"
	c.Mongo.DbName = ServiceName
	c.PurgeInterval = time.Hour
	c.PurgeAfter = 30 * 24 * time.Hour
}
```

Prefer `ServiceBindAddress` over the deprecated `ServiceBindHost`/`ServiceBindPort` pair, and
`SERVICE_INSTANCE_NAME` over the deprecated `SERVICE_NAME`. The default instance name is
`<ServiceName>(<hostname>)`.

## Secrets

- Type secrets as `config.HiddenString`. It is a string whose JSON (and therefore the YAML of
  `config.FormatEnv` and `/env`) is `****`. `config.ConnectionString` masks the password part of
  a URL instead.
- An **empty default** for a secret, plus a `Validate` rule that rejects it, makes the service
  fail closed instead of starting with a known key. Never hard-code a default key in `SetDefaults`.
- `.String()` gives the raw value where you need it, for example `auth.ApiKeyMiddleware(header, c.AdminAPIKey.String())`.

<!-- source: examples/notes/pkg/config/config.go -->
```go
AdminAPIKey config.HiddenString `mapstructure:"ADMIN_API_KEY" description:"API key expected in the X-API-Key header by /notes/env. Required: the service refuses to start without it."`

func (c *Config) Validate() error {
	if c.AdminAPIKey == "" {
		return errors.New("ADMIN_API_KEY is required")
	}
	...
}
```

## Validation

`Validate` is for cross-field rules and value checks that gost cannot know: required secrets,
positive durations, enum values, URL shapes, port collisions between two listeners. Return a
`fmt.Errorf` that names the env variable so the fatal log is actionable. Soft advice belongs in
`log.Warn`, not in an error.

## Generating `config.md`

Every leaf field carries a `description:"..."` tag; `description:"-"` hides a field. Two ways to
render the table:

1. At runtime: start the service once with `CONFIG_DESCRIPTION_PATH=config.md`.
2. In a test, without starting anything (the example does this and fails when the file is stale):

<!-- source: examples/notes/pkg/config/config_test.go -->
```go
var cfg config.Config
require.NoError(t, gost.SetDefaults(&cfg))

var buf bytes.Buffer
require.NoError(t, gost.WriteConfigDescription(&cfg, "", gost.NewMarkdownFormatter(&buf, true)))

path := filepath.Join("..", "..", "config.md")
if os.Getenv("UPDATE_CONFIG_MD") != "" {
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o644))

	return
}
want, err := os.ReadFile(path)
require.NoError(t, err)
require.Equal(t, string(want), buf.String(), "config.md is stale: run `make config-md`")
```

Commit exactly one `config.md` per service. The generated file for the example is
[`examples/notes/config.md`](../examples/notes/config.md); it also documents all gost variables
your service inherits.

## Using config in tests

`ReadConfig` is fatal on error, so tests use the non-fatal pieces:

```go
var cfg config.Config
require.NoError(t, gost.SetDefaults(&cfg))
cfg.AdminAPIKey = "test-key"
require.NoError(t, gost.Validate(&cfg))
```

`config.ReadFromEnv(&cfg, prefix)` is available when a test needs to exercise env parsing.

## Runtime mutable settings

Values that must change without a restart (log level, GC percent, batch sizes) are not config
fields but feature flags; see [Feature flags](feature-flags.md).
