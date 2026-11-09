# Service discovery

The `discovery` package registers the running instance and resolves other services by name.
Three backends share one interface; the backend is chosen by configuration, so a service does
not change when it moves from a single VM to a cluster.

```go
type Discovery interface {
	server.AsyncRunner                                        // a discovery is a runner: register it
	Services(ctx) ([]*ServiceInfo, error)
	ServiceByName(ctx, name string) (*ServiceInfo, error)
	SetMetadata(key, value string)                            // before Run
}
```

## Configuration

Embed `common.Config` under the `DISCOVERY` prefix:

<!-- source: examples/notes/pkg/config/config.go -->
```go
Discovery common.Config `mapstructure:"DISCOVERY"`
```

| Variable | Default | Meaning |
|---|---|---|
| `DISCOVERY_TYPE` | `none` | `none`, `in-house` or `consul` |
| `DISCOVERY_INHOUSE_EMBEDDED` | `true` | Gossip over the HTTP port through a WebSocket transport |
| `DISCOVERY_INHOUSE_BIND_ADDRESS` | `0.0.0.0:3001` | Gossip address when not embedded |
| `DISCOVERY_INHOUSE_CLUSTER_MEMBERS` | empty | Comma separated seed members, e.g. `10.0.0.5:3000` |
| `DISCOVERY_INHOUSE_SECRET_KEY` | a hard-coded key | **Override in production.** Must be 16, 24 or 32 bytes |
| `DISCOVERY_INHOUSE_DEREGISTER_SERVICE_AFTER` | `1h` | Remove a downed instance after this |
| `DISCOVERY_CONSUL_ADDRESS` | `http://127.0.0.1:8500` | |
| `DISCOVERY_CONSUL_TOKEN` | empty | ACL token |
| `DISCOVERY_CONSUL_HEALTH_CHECK_TTL` | `30s` | TTL check; the instance heartbeats every TTL/3 |
| `DISCOVERY_CONSUL_DEREREGISTER_TTL` | `30s` | (sic) critical for this long → deregistered |

The instance advertises `SERVICE_ADDRESS` (defaults to the bind address), the version, commit
and branch from the `version` package, and any `SetMetadata` pairs.

## Wiring

<!-- source: examples/notes/pkg/microservice/microservice.go -->
```go
discovery: common.NewDiscovery(cfg.Server, cfg.Discovery),   // in New

func (m *Microservice) BuildRoutes(r server.Router) {
	if d, ok := m.discovery.(*inhouse.Discovery); ok {
		d.Mount(r)                                            // embedded gossip transport endpoint
	}
	...
}

server.NewAsyncJobGroup(m.purger, m.discovery)                // in GetAsyncRunners
```

`DISCOVERY_TYPE=none` yields `fake.NewDiscovery`, which knows only the local instance (and
anything you `AddService` in tests). The in-house backend uses gossip (a fork of `hashicorp/memberlist`);
the Consul backend registers with a TTL check and skips registration when `SERVICE_ADDRESS` is
empty.

## Calling other services

`discovery.NewHttpClient(d)` returns an `*http.Client` whose transport treats the **URL host as a
service name**: `http://geoip/country/1.1.1.1` resolves `geoip`, round-robins its healthy
instances, and retries idempotent methods on another instance. `NewHttpClient(d, true)` is
sticky per host. Lower-level pieces: `discovery.NewTransport(base, d)`, `discovery.NewResolver(d)`
(5 minute cache, round-robin), `discovery.DefaultDialer(d)`, `discovery.GrpcDialer(d)`.

## Reacting to topology changes

Backends that implement `discovery.NotifyingDiscovery` accept subscriptions:

```go
d.Subscribe(discovery.NewEventHandler(func(ctx context.Context, inst discovery.ServiceInstanceInfo) {
	rebalance(inst)
}).ServiceName("worker").EventType(discovery.EventTypeUp, discovery.EventTypeDown))
```

Event types: `discovered`, `up`, `down`, `removed`. A distributed job manager can be built on
these events plus a consistent hash of instance IDs.
