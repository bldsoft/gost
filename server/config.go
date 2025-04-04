package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
)

const serviceInstanceAuto = "auto"

type Config struct {
	ServiceName               string `description:"-"` // General name of the service.
	DeprecatedServiceInstance string `mapstructure:"SERVICE_NAME" description:"DEPRECATED. Unique service instance name. Use 'auto' to set the hostname+service value. "`
	ServiceInstance           string `mapstructure:"SERVICE_INSTANCE_NAME" description:"Unique service instance name. Use 'auto' to set the hostname+service value. The name is used to identify the service in logs."`

	ServiceBindHost         string         `mapstructure:"SERVICE_HOST" description:"DEPRECATED. IP address, or a host name that can be resolved to IP addresses"`
	ServiceBindPort         int            `mapstructure:"SERVICE_PORT" description:"DEPRECATED. Service port"`
	ServiceBindAddressHTTP  config.Address `mapstructure:"SERVICE_BIND_ADDRESS" description:"Service configuration related to what address bind to and port to listen on for HTTP"`
	ServiceAddressHTTP      config.Address `mapstructure:"SERVICE_ADDRESS" description:"Service public address for HTTP"`
	ServiceBindAddressHTTPS config.Address `mapstructure:"SERVICE_BIND_ADDRESS_HTTPS" description:"Service configuration related to what address bind to and port to listen on for HTTPS"`
	ServiceAddressHTTPS     config.Address `mapstructure:"SERVICE_ADDRESS" description:"Service public address for HTTPS"`

	TLSCertificatePath string `mapstructure:"TLS_CERTIFICATE_PATH" description:"Path to TLS certificate file"`
	TLSKeyPath         string `mapstructure:"TLS_KEY_PATH" description:"Path to TLS key file"`
}

func (c *Config) LogExporterConfig() log.LogExporterConfig {
	return log.LogExporterConfig{
		Service:  c.ServiceName,
		Instance: c.ServiceInstance,
	}
}

func (c *Config) ServiceID() string {
	return c.ServiceInstance
}

// SetDefaults ...
func (c *Config) SetDefaults() {
	c.DeprecatedServiceInstance = serviceInstanceAuto
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
	if len(c.ServiceBindAddressHTTP) == 0 {
		log.Warn("SERVICE_HOST and SERVICE_PORT are deprecated, use SERVICE_BIND_ADDRESS instead")
		c.ServiceBindAddressHTTP = config.Address(net.JoinHostPort(c.ServiceBindHost, strconv.Itoa(c.ServiceBindPort)))
	}
	if len(c.ServiceAddressHTTP) == 0 {
		c.ServiceAddressHTTP = c.ServiceBindAddressHTTP
	}
	if c.DeprecatedServiceInstance == serviceInstanceAuto {
		c.DeprecatedServiceInstance = fmt.Sprintf("%s(%s)", c.ServiceName, utils.Hostname())
	}
	if len(c.ServiceInstance) == 0 {
		log.Warn("SERVICE_NAME is deprecated, use SERVICE_INSTANCE_NAME instead")
		c.ServiceInstance = c.DeprecatedServiceInstance
	}
	return err
}
