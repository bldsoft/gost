package inhouse

import (
	"github.com/bldsoft/gost/config"
)

type Config struct {
	Embedded       bool           `mapstructure:"EMBEDDED" description:"If true, in-house discovery will use service bind address"`
	BindAddress    config.Address `mapstructure:"BIND_ADDRESS" description:"For non embedded mode. Configuration related to what address to bind to and ports to listen on."`
	ClusterMembers []string       `mapstructure:"CLUSTER_MEMBERS" description:"Comma separated list of any existing member of the cluster to join it. Example: '127.0.0.1:3001'"`
}

func (c *Config) SetDefaults() {
	c.Embedded = true
	c.BindAddress = "0.0.0.0:3001"
}

func (c *Config) Validate() error {
	return nil
}
