package consul

import (
	"errors"
	"time"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/utils"
)

type ConsulConfig struct {
	ConsulAddr   string              `mapstructure:"ADDRESS" description:"Address of the Consul server"`
	ConsulScheme string              `mapstructure:"SCHEME" description:"URI scheme for the Consul server"`
	Token        config.HiddenString `mapstructure:"TOKEN" description:" Token is used to provide a per-request ACL token"`
}

func (c *ConsulConfig) SetDefaults() {
	c.ConsulScheme = "http"
	c.ConsulAddr = "127.0.0.1:8500"
}

func (c *ConsulConfig) Validate() error {
	return nil
}

type ServiceConfig struct {
	ServiceID      string        `mapstructure:"SERVICE_ID" description:"The ID of the service. If empty, a random one will be generated"`
	Cluster        string        `mapstructure:"CLUSTER" description:"The name of the service to register"`
	ServiceAddr    string        `mapstructure:"SERVICE_ADDRESS" description:"The address of the service"`
	ServicePort    int           `mapstructure:"SERVICE_PORT" description:"The port of the service"`
	HealthCheckTTL time.Duration `mapstructure:"HEALTH_CHECK_TTL" description:"Check TTL"`
	DeregisterTTL  time.Duration `mapstructure:"DEREREGISTER_TTL" description:"If a check is in the critical state for more than this configured value,	then the service will automatically be deregistered"`
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
		c.ServiceID = utils.RandString(32)
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
