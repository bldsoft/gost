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

	ServiceBindHost    string         `mapstructure:"SERVICE_HOST" description:"DEPRECATED. IP address, or a host name that can be resolved to IP addresses"`
	ServiceBindPort    int            `mapstructure:"SERVICE_PORT" description:"DEPRECATED. Service port"`
	ServiceBindAddress config.Address `mapstructure:"SERVICE_BIND_ADDRESS" description:"Service configuration related to what address bind to and port to listen on for HTTP"`
	ServiceAddress     config.Address `mapstructure:"SERVICE_ADDRESS" description:"Service public address for HTTP"`

	TLS TLSConfig `mapstructure:"TLS"`
}

type TLSConfig struct {
	ServiceBindAddress config.Address `mapstructure:"SERVICE_BIND_ADDRESS" description:"Service configuration related to what address bind to and port to listen on for HTTPS"`
	ServiceAddress     config.Address `mapstructure:"SERVICE_ADDRESS" description:"Service public address for HTTPS"`

	CertificatePath string `mapstructure:"CERTIFICATE_PATH" description:"Path to TLS certificate file"`
	KeyPath         string `mapstructure:"KEY_PATH" description:"Path to TLS key file"`
}

func (tls *TLSConfig) Validate(parent *Config) error {
	if len(tls.CertificatePath) == 0 {
		return nil
	}

	if len(tls.ServiceAddress) == 0 {
		tls.ServiceAddress = parent.ServiceAddress
	}
	if len(tls.ServiceBindAddress) == 0 {
		tls.ServiceBindAddress = config.Address(net.JoinHostPort(parent.ServiceBindAddress.Host(), "443"))
	}
	return nil
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

	c.ServiceBindAddress = config.Address("0.0.0.0:3000")
}

// Validate ...
func (c *Config) Validate() error {
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
	if c.DeprecatedServiceInstance == serviceInstanceAuto {
		c.DeprecatedServiceInstance = fmt.Sprintf("%s(%s)", c.ServiceName, utils.Hostname())
	}
	if len(c.ServiceInstance) == 0 {
		log.Warn("SERVICE_NAME is deprecated, use SERVICE_INSTANCE_NAME instead")
		c.ServiceInstance = c.DeprecatedServiceInstance
	}
	return c.TLS.Validate(c)
}
