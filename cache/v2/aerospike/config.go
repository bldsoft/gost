package aerospike

type Config struct {
	Host      string `mapstructure:"HOSTS"`
	Port      int    `mapstructure:"PORT"`
	KeyPrefix string `mapstructure:"KEY_PREFIX"`
	Namespace string `mapstructure:"NAMESPACE"`

	ConnectionPolicy ConnectionPolicy `mapstructure:"CONNECTION_POLICY"`
	WritePolicy      ReadWritePolicy  `mapstructure:"WRITE_POLICY"`
	ReadPolicy       ReadWritePolicy  `mapstructure:"READ_POLICY"`

	// TODO: add support for the fields if needed
	// ClusterName string `mapstructure:"CLUSTER_NAME"`
	// Username    string `mapstructure:"USERNAME"`
	// Password    string `mapstructure:"PASSWORD"`

}

type ConnectionPolicy struct {
	ConnectionQueueSize   int `mapstructure:"CONNECTION_QUEUE_SIZE"`
	TimeOutMs             int `mapstructure:"TIMEOUT_MS"`
	IdleTimeoutMs         int `mapstructure:"IDLE_TIMEOUT_MS"`
	MinConnectionsPerNode int `mapstructure:"MIN_CONNECTIONS_PER_NODE"`
}

type ReadWritePolicy struct {
	TotalTimeoutMs        int `mapstructure:"TOTAL_TIMEOUT_MS"`
	MaxRetries            int `mapstructure:"MAX_RETRIES"`
	SleepBetweenRetriesMs int `mapstructure:"SLEEP_BETWEEN_RETRIES_MS"`
	SocketTimeoutMs       int `mapstructure:"SOCKET_TIMEOUT_MS"`
}

func (c *Config) SetDefaults() {
	c.Host = "127.0.0.1"
	c.Port = 3000

	c.ConnectionPolicy.ConnectionQueueSize = 2000
	c.ConnectionPolicy.TimeOutMs = 500
	c.ConnectionPolicy.IdleTimeoutMs = 30

	c.WritePolicy.TotalTimeoutMs = 5
	c.WritePolicy.MaxRetries = 2
	c.WritePolicy.SleepBetweenRetriesMs = 1
	c.WritePolicy.SocketTimeoutMs = 5

	c.ReadPolicy.TotalTimeoutMs = 5
	c.ReadPolicy.MaxRetries = 2
	c.ReadPolicy.SleepBetweenRetriesMs = 1
	c.ReadPolicy.SocketTimeoutMs = 5
}
