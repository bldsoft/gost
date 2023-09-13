package inhouse

import (
	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery"
)

type InHouseConfig struct {
	Address        config.Address `mapstructure:"ADDRESS" description:"The address used for in-house discovery communication"`
	ClusterMembers []string       `mapstructure:"CLUSTER_MEMBERS" description:"Comma separated list of any existing member of the cluster to join it. Example: '127.0.0.1:3001'"`
}

func (c *InHouseConfig) SetDefaults() {
	c.Address = "0.0.0.0:3001"
}

func (c *InHouseConfig) Validate() error { return nil }

type Config struct {
	InHouseConfig           `mapstructure:"DISCOVERY_INHOUSE"`
	discovery.ServiceConfig `mapstructure:"DISCOVERY"`
}
