package inhouse

import (
	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery"
)

type InHouseConfig struct {
	BindAddress      config.Address `mapstructure:"BIND_ADDRESS" description:"Configuration related to what address to bind to and ports to listen on."`
	AdvertiseAddress config.Address `mapstructure:"ADDRESS" description:"Configuration related to what address to advertise to other cluster members. The address used for in-house discovery communication"`
	ClusterMembers   []string       `mapstructure:"CLUSTER_MEMBERS" description:"Comma separated list of any existing member of the cluster to join it. Example: '127.0.0.1:3001'"`
}

func (c *InHouseConfig) SetDefaults() {
	c.BindAddress = "0.0.0.0:3001"
	c.AdvertiseAddress = c.BindAddress
}

func (c *InHouseConfig) Validate() error {
	return nil
}

type Config struct {
	InHouseConfig           `mapstructure:"DISCOVERY_INHOUSE"`
	discovery.ServiceConfig `mapstructure:"DISCOVERY"`
}
