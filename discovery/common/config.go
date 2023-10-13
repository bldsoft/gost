package common

import (
	"github.com/bldsoft/gost/discovery/consul"
	"github.com/bldsoft/gost/discovery/inhouse"
)

//go:generate go run github.com/abice/go-enum -f=$GOFILE

// ENUM(none, in-house, consul)
type DiscoveryType string

type Config struct {
	DiscoveryType DiscoveryType  `mapstructure:"TYPE" description:"Discovery type (none, in-house, consul)"`
	InHouse       inhouse.Config `mapstructure:"INHOUSE"`
	Consul        consul.Config  `mapstructure:"CONSUL"`
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
