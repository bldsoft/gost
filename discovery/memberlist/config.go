package memberlist

import "github.com/bldsoft/gost/discovery"

type MemberListConfig struct {
	MemberlistAddr string   `mapstructure:"ADDRESS" description:"The address of the service. If it's empty the service doesn't register in consul"`
	MemberlistPort int      `mapstructure:"PORT" description:"Any existing member of the cluster to join it"`
	ClusterMembers []string `mapstructure:"CLUSTER_MEMBERS" description:"Any existing member of the cluster to join it"`
}

func (c *MemberListConfig) SetDefaults() {}

func (c *MemberListConfig) Validate() error { return nil }

type Config struct {
	MemberListConfig        `mapstructure:"DISCOVERY_MEMBERLIST"`
	discovery.ServiceConfig `mapstructure:"DISCOVERY"`
}
