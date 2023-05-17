package memcached

type Config struct {
	Servers         []string `mapstructure:"SERVERS"`
	ReadOnlyServers []string `mapstructure:"READONLY_SERVERS" json:"READONLY_SERVERS" description:"alternate memcached instace for GET operations"`
	TimeoutMs       int      `mapstructure:"TIMEOUT_MS" json:"TIMEOUT_MS"`
	KeyPrefix       string   `mapstructure:"KEY_PREFIX" json:"KEY_PREFIX" description:"required for mcrouter to prevent prefix routing"`
	MaxIdleConns    *int     `mapstructure:"MAX_IDLE_CONNS" json:"MAX_IDLE_CONNS" description:"maximum number of idle connections to the memcached server"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Servers = []string{"172.17.0.1:11211"}
	c.TimeoutMs = 0
}

// Validate ...
func (c *Config) Validate() error {
	return nil
}
