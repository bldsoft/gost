package discovery

import (
	"errors"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
	"github.com/bldsoft/gost/version"
)

type ServiceConfig struct {
	ServiceID   string `mapstructure:"SERVICE_ID" description:"The ID of the service. This must be unique in the cluster. If empty, a random one will be generated"`
	ServiceName string `mapstructure:"SERVICE_NAME" description:"The name of the service to register"`
	ServiceAddr string `mapstructure:"SERVICE_ADDRESS" description:"The address of the service"`
	meta        map[string]string
}

func (c *ServiceConfig) SplittedAddr() (proto, host, port string) {
	host = c.ServiceAddr
	if i := strings.Index(c.ServiceAddr, "://"); i > 0 {
		proto = c.ServiceAddr[:i]
		host = c.ServiceAddr[i+3:]
	}
	if h, p, err := net.SplitHostPort(host); err == nil {
		host = h
		port = p
	}
	return proto, "", ""
}

func (c *ServiceConfig) Proto() string {
	proto, _, _ := c.SplittedAddr()
	return proto
}

func (c *ServiceConfig) Host() string {
	_, host, _ := c.SplittedAddr()
	return host
}

func (c *ServiceConfig) Port() int {
	_, _, port := c.SplittedAddr()
	portInt, _ := strconv.Atoi(port)
	return portInt
}

func (c *ServiceConfig) AddMetadata(key, value string) {
	c.meta[key] = value
}

func (c *ServiceConfig) SetDefaults() {
	c.meta = make(map[string]string)
	c.ServiceAddr = "http://0.0.0.0:3000"
}

func (c *ServiceConfig) Validate() error {
	if len(c.ServiceName) == 0 {
		return errors.New("SERVICE_NAME is required")
	}
	if len(c.ServiceID) == 0 {
		c.ServiceID = utils.RandString(32)
	}
	return nil
}

func (c *ServiceConfig) ServiceInstanceInfo() ServiceInstanceInfo {
	proto, host, port := c.SplittedAddr()
	return ServiceInstanceInfo{
		ID:      c.ServiceID,
		Host:    host,
		Proto:   proto,
		Port:    port,
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
