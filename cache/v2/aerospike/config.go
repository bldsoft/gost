package aerospike

type Config struct {
	Host      string `mapstructure:"HOSTS"`
	Port      int    `mapstructure:"PORT"`
	KeyPrefix string `mapstructure:"KEY_PREFIX"`

	ClusterName string `mapstructure:"CLUSTER_NAME"`
	Username    string `mapstructure:"USERNAME"`
	Password    string `mapstructure:"PASSWORD"`
	Namespace   string `mapstructure:"NAMESPACE"`
}

func (c *Config) SetDefaults() {
	c.Host = "172.17.0.1:3000"
	c.Port = 3000
}
