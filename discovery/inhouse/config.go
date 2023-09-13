package inhouse

import "github.com/bldsoft/gost/discovery"

type InHouseConfig struct {
	InHouseHost    string   `mapstructure:"HOST" description:"Host used for in-house discovery communication"`
	InHousePort    int      `mapstructure:"PORT" description:"Port used for in-house discovery communication"`
	ClusterMembers []string `mapstructure:"CLUSTER_MEMBERS" description:"Any existing member of the cluster to join it"`
}

func (c *InHouseConfig) SetDefaults() {}

func (c *InHouseConfig) Validate() error { return nil }

type Config struct {
	InHouseConfig           `mapstructure:"DISCOVERY_INHOUSE"`
	discovery.ServiceConfig `mapstructure:"DISCOVERY"`
}
