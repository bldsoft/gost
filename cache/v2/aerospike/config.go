package aerospike

type Config struct {
	Host      string `mapstructure:"HOSTS"`
	Port      int    `mapstructure:"PORT"`
	KeyPrefix string `mapstructure:"KEY_PREFIX"`

	// TODO: add support for the fields if needed
	// ClusterName string `mapstructure:"CLUSTER_NAME"`
	// Username    string `mapstructure:"USERNAME"`
	// Password    string `mapstructure:"PASSWORD"`
	// Namespace   string `mapstructure:"NAMESPACE"`
}

func (c *Config) SetDefaults() {
	c.Host = "127.0.0.1"
	c.Port = 3000
}
