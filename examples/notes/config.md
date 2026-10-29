|**Environment variable**|**Value**|**Description**|
|------------------------|---------|---------------|
|SERVICE_NAME|auto|DEPRECATED. Unique service instance name. Use 'auto' to set the hostname+service value. |
|SERVICE_INSTANCE_NAME||Unique service instance name. Use 'auto' to set the hostname+service value. The name is used to identify the service in logs.|
|SERVICE_HOST|0.0.0.0|DEPRECATED. IP address, or a host name that can be resolved to IP addresses|
|SERVICE_PORT|3000|DEPRECATED. Service port|
|SERVICE_BIND_ADDRESS|0.0.0.0:3000|Service configuration related to what address bind to and port to listen on for HTTP|
|SERVICE_ADDRESS||Service public address|
|PROBES_BIND_ADDRESS||Address to serve the Kubernetes probes on instead of the main listener. Empty keeps them on SERVICE_BIND_ADDRESS|
|TLS_SERVICE_BIND_ADDRESS||Service configuration related to what address bind to and port to listen on for HTTPS|
|TLS_CERTIFICATE_PATH||Path to TLS certificate file|
|TLS_KEY_PATH||Path to TLS key file|
|READ_TIMEOUT_MS|0|Maximum duration for reading the entire request, including the body (in milliseconds)|
|WRITE_TIMEOUT_MS|0|Maximum duration before timing out writes of the response (in milliseconds)|
|LOG_LEVEL|info|Log level|
|LOG_COLOR_ENABLED|true|Enable the colorized output|
|DATABASE_SERVER|mongodb://localhost:27017|MongoDB connection string|
|DATABASE_DBNAME|notes|Service database name|
|DATABASE_MIGRATION_COLLECTION|schema_migrations|MongoDB migration collection name|
|DISCOVERY_TYPE|none|Discovery type (none, in-house, consul)|
|DISCOVERY_INHOUSE_EMBEDDED|true|If true, in-house discovery will use service bind address|
|DISCOVERY_INHOUSE_BIND_ADDRESS|0.0.0.0:3001|For non embedded mode. Configuration related to what address to bind to and ports to listen on.|
|DISCOVERY_INHOUSE_CLUSTER_MEMBERS||Comma separated list of any existing member of the cluster to join it. Example: '127.0.0.1:3001'|
|DISCOVERY_INHOUSE_SECRET_KEY|ZljFlK6atNj5U3VbHrDxRgFMHYcgEOpy|SecretKey is used to encrypt messages. The value should be either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.|
|DISCOVERY_INHOUSE_DEREGISTER_SERVICE_AFTER|1h0m0s|The interval after which the downed service is removed from the cluster|
|DISCOVERY_CONSUL_ADDRESS|http://127.0.0.1:8500|Address of the Consul server|
|DISCOVERY_CONSUL_TOKEN|| Token is used to provide a per-request ACL token|
|DISCOVERY_CONSUL_HEALTH_CHECK_TTL|30s|Check TTL|
|DISCOVERY_CONSUL_DEREREGISTER_TTL|30s|If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered|
|ADMIN_API_KEY||API key expected in the X-API-Key header by the /env endpoint. Required: the service refuses to start without it.|
|NOTES_PURGE_INTERVAL|1h0m0s|How often soft-deleted notes are purged (Go duration, e.g. 1h)|
|NOTES_PURGE_AFTER|720h0m0s|Soft-deleted notes older than this are removed permanently (Go duration, e.g. 720h)|
