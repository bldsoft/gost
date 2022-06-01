package memcached

type Config struct {
	Servers []string `mapstructure:"SERVERS"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.Servers = []string{"172.17.0.1:11211"}
}

// Validate ...
func (c *Config) Validate() error {
	return nil
}
