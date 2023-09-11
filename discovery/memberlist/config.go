package memberlist

import "github.com/bldsoft/gost/discovery"

type MemberListConfig struct {
	MemberlistHost string   `mapstructure:"HOST" description:"Memberlist host"`
	MemberlistPort int      `mapstructure:"PORT" description:"Meberlist port"`
	ClusterMembers []string `mapstructure:"CLUSTER_MEMBERS" description:"Any existing member of the cluster to join it"`
}

func (c *MemberListConfig) SetDefaults() {}

func (c *MemberListConfig) Validate() error { return nil }

type Config struct {
	MemberListConfig        `mapstructure:"DISCOVERY_MEMBERLIST"`
	discovery.ServiceConfig `mapstructure:"DISCOVERY"`
}
