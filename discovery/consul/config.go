package consul

import (
	"time"

	"github.com/bldsoft/gost/config"
)

type Config struct {
	ConsulAddr     config.HttpAddress  `mapstructure:"ADDRESS" description:"Address of the Consul server"`
	Token          config.HiddenString `mapstructure:"TOKEN" description:" Token is used to provide a per-request ACL token"`
	HealthCheckTTL time.Duration       `mapstructure:"HEALTH_CHECK_TTL" description:"Check TTL"`
	DeregisterTTL  time.Duration       `mapstructure:"DEREREGISTER_TTL" description:"If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered"`
}

func (c *Config) SetDefaults() {
	c.ConsulAddr = "http://127.0.0.1:8500"
	c.HealthCheckTTL = 30 * time.Second
	c.DeregisterTTL = c.HealthCheckTTL
}

func (c *Config) Validate() error {
	return nil
}
