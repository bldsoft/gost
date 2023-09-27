package server

import (
	"errors"
	"net"
	"strconv"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
)

const serviceInstanceHostname = "hostname"

type Config struct {
	ServiceName               string `description:"-"` // General name of the service.
	DeprecatedServiceInstance string `mapstructure:"SERVICE_NAME" description:"DEPRECATED. Unique service instance name. Use 'hostname' to set the hostname value. "`
	ServiceInstance           string `mapstructure:"SERVICE_INSTANCE_NAME" description:"Unique service instance name. Use 'hostname' to set the hostname value. "`

	ServiceBindHost    string         `mapstructure:"SERVICE_HOST" description:"DEPRECATED. IP address, or a host name that can be resolved to IP addresses"`
	ServiceBindPort    int            `mapstructure:"SERVICE_PORT" description:"DEPRECATED. Service port"`
	ServiceBindAddress config.Address `mapstructure:"SERVICE_BIND_ADDRESS" description:"Service configuration related to what address bind to and port to listen on"`
	ServiceAddress     config.Address `mapstructure:"SERVICE_ADDRESS" description:"Service public address"`
}

func (c *Config) ServiceID() string {
	return c.ServiceInstance
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.DeprecatedServiceInstance = serviceInstanceHostname
	// c.ServiceInstance = serviceInstanceHostname // use value from DeprecatedServiceInstance
	c.ServiceBindHost = "0.0.0.0"
	c.ServiceBindPort = 3000
}

// Validate ...
func (c *Config) Validate() error {
	var err error
	if len(c.ServiceName) == 0 {
		return errors.New("ServiceName is not set. Do it in SetDefaults method")
	}
	if len(c.ServiceBindAddress) == 0 {
		log.Warn("SERVICE_HOST and SERVICE_PORT are deprecated, use SERVICE_BIND_ADDRESS instead")
		c.ServiceBindAddress = config.Address(net.JoinHostPort(c.ServiceBindHost, strconv.Itoa(c.ServiceBindPort)))
	}
	if len(c.ServiceAddress) == 0 {
		c.ServiceAddress = c.ServiceBindAddress
	}
	if c.DeprecatedServiceInstance == serviceInstanceHostname {
		c.DeprecatedServiceInstance = utils.Hostname()
	}
	if len(c.ServiceInstance) == 0 {
		log.Warn("SERVICE_NAME is deprecated, use SERVICE_INSTANCE_NAME instead")
		c.ServiceInstance = c.DeprecatedServiceInstance
	}
	return err
}
