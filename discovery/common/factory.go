package common

import (
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/discovery/consul"
	"github.com/bldsoft/gost/discovery/fake"
	"github.com/bldsoft/gost/discovery/inhouse"
)

func NewDiscovery(cfg Config) discovery.Discovery {
	switch cfg.DiscoveryType {
	case DiscoveryTypeConsul:
		return consul.NewDiscovery(consul.Config{
			ConsulConfig:  cfg.Consul,
			ServiceConfig: cfg.ServiceConfig,
		})
	case DiscoveryTypeInHouse:
		return inhouse.NewDiscovery(inhouse.Config{
			InHouseConfig: cfg.InHouse,
			ServiceConfig: cfg.ServiceConfig,
		})
	case DiscoveryTypeNone:
		fallthrough
	default:
		return fake.NewDiscovery(cfg.ServiceConfig)
	}
}
