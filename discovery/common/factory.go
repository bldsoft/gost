package common

import (
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/discovery/consul"
	"github.com/bldsoft/gost/discovery/memberlist"
)

func NewDiscovery(cfg Config) discovery.Discovery {
	switch cfg.DiscoveryType {
	case DiscoveryTypeConsul:
		return consul.NewDiscovery(consul.Config{
			ConsulConfig:  cfg.Consul,
			ServiceConfig: cfg.ServiceConfig,
		})
	case DiscoveryTypeMemberlist:
		fallthrough
	default:
		return memberlist.NewDiscovery(memberlist.Config{
			MemberListConfig: cfg.MemberList,
			ServiceConfig:    cfg.ServiceConfig,
		})
	}
}
