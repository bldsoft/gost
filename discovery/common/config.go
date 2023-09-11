package common

import (
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/discovery/consul"
	"github.com/bldsoft/gost/discovery/memberlist"
)

//go:generate go run github.com/abice/go-enum -f=$GOFILE

// ENUM(none, memberlist, consul)
type DiscoveryType string

type Config struct {
	discovery.ServiceConfig
	DiscoveryType DiscoveryType               `mapstructure:"TYPE" description:"Discovery type"`
	MemberList    memberlist.MemberListConfig `mapstructure:"MEMBERLIST"`
	Consul        consul.ConsulConfig         `mapstructure:"CONSUL"`
}

func (c *Config) SetDefaults() {
	c.DiscoveryType = DiscoveryTypeNone
}

func (c *Config) Validate() error {
	if _, err := ParseDiscoveryType(string(c.DiscoveryType)); err != nil {
		return err
	}
	return nil
}
