package common

import (
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/discovery/consul"
	"github.com/bldsoft/gost/discovery/memberlist"
)

//go:generate go run github.com/abice/go-enum -f=$GOFILE --noprefix

// ENUM(memberlist, consul)
type DiscoveryType string

type Config struct {
	DiscoveryType           DiscoveryType `mapstructure:"DISCOVERY_TYPE" description:"Discovery type"`
	discovery.ServiceConfig `mapstructure:"DISCOVERY"`
	MemberList              memberlist.MemberListConfig `mapstructure:"DISCOVERY_MEMBERLIST"`
	Consul                  consul.ConsulConfig         `mapstructure:"DISCOVERY_CONSUL"`
}

func (c *Config) SetDefaults() {
	c.DiscoveryType = DiscoveryTypeMemberlist
}

func (c *Config) Validate() error {
	if _, err := ParseDiscoveryType(string(c.DiscoveryType)); err != nil {
		return err
	}
	return nil
}
