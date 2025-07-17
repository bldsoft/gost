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
	ConnectionQueueSize int  `mapstructure:"CONNECTION_QUEUE_SIZE"`
	TimeOutMs           int  `mapstructure:"TIMEOUT_MS"`
	IdleTimeoutMs       int  `mapstructure:"IDLE_TIMEOUT_MS"`
	FailIfNotConnected  bool `mapstructure:"FAIL_IF_NOT_CONNECTED"`
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
}
