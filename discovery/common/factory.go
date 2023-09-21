package common

import (
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/discovery/consul"
	"github.com/bldsoft/gost/discovery/fake"
	"github.com/bldsoft/gost/discovery/inhouse"
	"github.com/bldsoft/gost/server"
)

func NewDiscovery(serviceCfg server.Config, cfg Config) discovery.Discovery {
	switch cfg.DiscoveryType {
	case DiscoveryTypeConsul:
		return consul.NewDiscovery(serviceCfg, cfg.Consul)
	case DiscoveryTypeInHouse:
		return inhouse.NewDiscovery(serviceCfg, cfg.InHouse)
	case DiscoveryTypeNone:
		fallthrough
	default:
		return fake.NewDiscovery(serviceCfg)
	}
}
