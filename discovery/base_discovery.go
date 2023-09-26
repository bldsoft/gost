package discovery

import (
	"context"

	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/utils"
	"github.com/bldsoft/gost/version"
)

type BaseDiscovery struct {
	ServiceInfo ServiceInstanceInfo

	handlersByEventyType [EventTypeCount][]*EventHandler
}

func NewBaseDiscovery(serviceCfg server.Config) BaseDiscovery {
	return BaseDiscovery{
		ServiceInfo: ServiceInstanceInfo{
			ServiceName: serviceCfg.ServiceName,
			ID:          serviceCfg.ServiceID(),
			Address:     serviceCfg.ServiceAddress,
			Node:        utils.Hostname(),
			Version:     version.Version,
			Commit:      version.GitCommit,
			Branch:      version.GitBranch,
			Healthy:     true,
			Meta:        make(map[string]string),
		},
	}
}

// must be called before discovery run
func (d *BaseDiscovery) SetMetadata(key, value string) {
	d.ServiceInfo.Meta[key] = value
}

func (d *BaseDiscovery) subscribe(handler *EventHandler) {
	if handler == nil {
		return
	}
	for _, t := range handler.EventTypes() {
		d.handlersByEventyType[t] = append(d.handlersByEventyType[t], handler)
	}
}

func (d *BaseDiscovery) Subscribe(handler *EventHandler, handlers ...*EventHandler) {
	d.subscribe(handler)
	for _, h := range handlers {
		d.subscribe(h)
	}
}

func (d *BaseDiscovery) TriggerEventCtx(ctx context.Context, eventType EventType, instance ServiceInstanceInfo) {
	for _, handler := range d.handlersByEventyType[eventType] {
		handler.TriggerEvent(ctx, instance)
	}
}

func (d *BaseDiscovery) TriggerEvent(eventType EventType, instance ServiceInstanceInfo) {
	d.TriggerEventCtx(context.Background(), eventType, instance)
}
