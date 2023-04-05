package consul

import (
	"errors"
	"time"
)

type ConsulConfig struct {
	ConsulAddr   string `mapstructure:"ADDRESS"`
	ConsulScheme string `mapstructure:"SCHEME"`
}

func (c *ConsulConfig) SetDefaults() {
	c.ConsulScheme = "http"
	c.ConsulAddr = "127.0.0.1:8500"
}

func (c *ConsulConfig) Validate() error {
	return nil
}

type ServiceConfig struct {
	ServiceID      string        `mapstructure:"SERVICE_ID"`
	Cluster        string        `mapstructure:"CLUSTER"`
	ServiceAddr    string        `mapstructure:"SERVICE_ADDRESS"`
	ServicePort    int           `mapstructure:"SERVICE_PORT"`
	HealthCheckTTL time.Duration `mapstructure:"HEALTH_CHECK_TTL"`
	DeregisterTTL  time.Duration `mapstructure:"DEREREGISTER_TTL"`
}

func (c *ServiceConfig) checkID() string {
	return c.ServiceID
}

func (c *ServiceConfig) SetDefaults() {
	c.HealthCheckTTL = 30 * time.Second
	c.DeregisterTTL = c.HealthCheckTTL
}

func (c *ServiceConfig) Validate() error {
	if len(c.Cluster) == 0 {
		return errors.New("CONSUL_CLUSTER is required")
	}
	if len(c.ServiceID) == 0 {
		return errors.New("CONSUL_SERVICE_ID is required")
	}
	return nil
}

type Config struct {
	ConsulConfig  `mapstructure:"CONSUL"`
	ServiceConfig `mapstructure:"CONSUL"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {}

// Validate ...
func (c *Config) Validate() error {
	var err error

	return err
}
