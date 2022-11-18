package memcached

type Config struct {
	Servers []string `mapstructure:"SERVERS"`
	TimeoutMs int `mapstructure:"TIMEOUT_MS"`
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
