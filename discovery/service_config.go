package discovery

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
	"github.com/bldsoft/gost/version"
)

type ServiceConfig struct {
	ServiceID    string `mapstructure:"SERVICE_ID" description:"The ID of the service. This must be unique in the cluster. If empty, a random one will be generated"`
	ServiceName  string `mapstructure:"SERVICE_NAME" description:"The name of the service to register"`
	ServiceProto string `mapstructure:"SERVICE_PROTO" description:"The proto of the service"`
	ServiceHost  string `mapstructure:"SERVICE_HOST" description:"The address of the service. If it's empty the service doesn't register in discovery"`
	ServicePort  string `mapstructure:"SERVICE_PORT" description:"The port of the service"`
	meta         map[string]string
}

func (c *ServiceConfig) Port() int {
	port, _ := strconv.Atoi(c.ServicePort)
	return port
}

func (c *ServiceConfig) AddMetadata(key, value string) {
	c.meta[key] = value
}

func (c *ServiceConfig) SetDefaults() {
	c.meta = make(map[string]string)
	c.ServiceProto = "http"
}

func (c *ServiceConfig) Validate() error {
	if len(c.ServiceName) == 0 {
		return errors.New("SERVICE_NAME is required")
	}
	if len(c.ServiceID) == 0 {
		c.ServiceID = utils.RandString(32)
	}

	if c.ServicePort != "" {
		if _, err := strconv.Atoi(c.ServicePort); err != nil {
			return fmt.Errorf("failed to parse port: %w", err)
		}
	}
	return nil
}

func (c *ServiceConfig) ServiceInstanceInfo() ServiceInstanceInfo {
	return ServiceInstanceInfo{
		Host:    c.ServiceHost,
		Proto:   c.ServiceProto,
		Port:    c.ServicePort,
		Node:    Hostname(),
		Version: version.Version,
		Commit:  version.GitCommit,
		Branch:  version.GitBranch,
		Healthy: true,
		Meta:    c.meta,
	}
}

func Hostname(allowPanic ...bool) string {
	hostname, err := os.Hostname()
	if err != nil {
		if len(allowPanic) > 0 || allowPanic[0] {
			panic(err)
		}
		log.Errorf("Failed to get hostname: %s", err)
	}
	return hostname
}
