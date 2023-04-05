package consul

import (
	"errors"
	"time"
)

type Config struct {
	ConsulAddr     string        `mapstructure:"CONSUL_ADDRESS"`
	ConsulScheme   string        `mapstructure:"CONSUL_SCHEME"`
	ServiceID      string        `mapstructure:"CONSUL_SERVICE_ID"`
	Cluster        string        `mapstructure:"CONSUL_CLUSTER"`
	ServiceAddr    string        `mapstructure:"CONSUL_SERVICE_ADDRESS"`
	ServicePort    int           `mapstructure:"CONSUL_SERVICE_PORT"`
	HealthCheckTTL time.Duration `mapstructure:"CONSUL_HEALTH_CHECK_TTL"`
	DeregisterTTL  time.Duration `mapstructure:"CONSUL_DEREREGISTER_TTL"`
}

func (c *Config) checkID() string {
	return c.ServiceID
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.ConsulScheme = "http"
	c.ConsulAddr = "127.0.0.1:8600"
	c.HealthCheckTTL = 30 * time.Second
	c.DeregisterTTL = 3 * 24 * time.Hour
}

// Validate ...
func (c *Config) Validate() error {
	var err error
	if len(c.Cluster) == 0 {
		return errors.New("CONSUL_CLUSTER is required")
	}
	if len(c.ServiceID) == 0 {
		return errors.New("CONSUL_SERVICE_ID is required")
	}
	return err
}
