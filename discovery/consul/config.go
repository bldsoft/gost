package consul

import (
	"time"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery"
)

type ConsulConfig struct {
	ConsulAddr     string              `mapstructure:"ADDRESS" description:"Address of the Consul server"`
	ConsulScheme   string              `mapstructure:"SCHEME" description:"URI scheme for the Consul server"`
	Token          config.HiddenString `mapstructure:"TOKEN" description:" Token is used to provide a per-request ACL token"`
	HealthCheckTTL time.Duration       `mapstructure:"HEALTH_CHECK_TTL" description:"Check TTL"`
	DeregisterTTL  time.Duration       `mapstructure:"DEREREGISTER_TTL" description:"If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered"`
}

func (c *ConsulConfig) SetDefaults() {
	c.ConsulScheme = "http"
	c.ConsulAddr = "127.0.0.1:8500"
	c.HealthCheckTTL = 30 * time.Second
	c.DeregisterTTL = c.HealthCheckTTL
}

func (c *ConsulConfig) Validate() error {
	return nil
}

type Config struct {
	ConsulConfig            `mapstructure:"DISCOVERY_CONSUL"`
	discovery.ServiceConfig `mapstructure:"DISCOVERY"`
}

// SetDefaults ...
func (c *Config) SetDefaults() {}

// Validate ...
func (c *Config) Validate() error {
	var err error

	return err
}
