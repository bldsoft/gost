package fake

import (
	"context"
	"sort"
	"sync"

	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/server"
)

type Discovery struct {
	discovery.BaseDiscovery
	server.AsyncJob

	services    map[string]*discovery.ServiceInfo
	servicesMtx sync.RWMutex
}

func NewDiscovery(cfg server.Config) *Discovery {
	d := &Discovery{BaseDiscovery: discovery.NewBaseDiscovery(cfg), services: make(map[string]*discovery.ServiceInfo)}
	d.services[cfg.ServiceName] = &discovery.ServiceInfo{
		Name:      cfg.ServiceName,
		Instances: []discovery.ServiceInstanceInfo{d.ServiceInfo},
	}
	return d
}

func (d *Discovery) AddService(serviceInfo *discovery.ServiceInfo) {
	d.servicesMtx.Lock()
	defer d.servicesMtx.Unlock()
	d.services[serviceInfo.Name] = serviceInfo
	for _, instance := range serviceInfo.Instances {
		d.TriggerEvent(discovery.EventTypeDiscovered, instance)
		d.TriggerEvent(discovery.EventTypeUp, instance)
	}
}

func (d *Discovery) Services(ctx context.Context) ([]*discovery.ServiceInfo, error) {
	d.servicesMtx.RLock()
	defer d.servicesMtx.RUnlock()
	res := make([]*discovery.ServiceInfo, 0, len(d.services))
	for _, s := range d.services {
		res = append(res, s)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].Name < res[j].Name
	})
	return res, nil
}

func (d *Discovery) ServiceByName(ctx context.Context, name string) (*discovery.ServiceInfo, error) {
	d.servicesMtx.RLock()
	defer d.servicesMtx.RUnlock()
	s, ok := d.services[name]
	if !ok {
		return nil, discovery.NotFound
	}
	return s, nil
}

var _ discovery.NotifyingDiscovery = &Discovery{}
