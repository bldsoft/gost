package discovery

import (
	"context"
	"os"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/server"
	"github.com/bldsoft/gost/version"
)

type BaseDiscovery struct {
	ServiceInfo ServiceInstanceInfo

	handlersByEventyType [ServiceEventTypeCount][]EventServiceHandler
}

func NewBaseDiscovery(serviceCfg server.Config) BaseDiscovery {
	return BaseDiscovery{
		ServiceInfo: ServiceInstanceInfo{
			ServiceName: serviceCfg.ServiceName,
			ID:          serviceCfg.ServiceID(),
			Address:     serviceCfg.ServiceAddress,
			Node:        Hostname(),
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

func (d *BaseDiscovery) subscribe(handler EventServiceHandler) {
	for _, t := range handler.Types() {
		d.handlersByEventyType[t] = append(d.handlersByEventyType[t], handler)
	}
}

func (d *BaseDiscovery) Subscribe(handler EventServiceHandler, handlers ...EventServiceHandler) {
	d.subscribe(handler)
	for _, h := range handlers {
		d.subscribe(h)
	}
}

func (d *BaseDiscovery) TriggerEventCtx(ctx context.Context, eventType ServiceEventType, instance ServiceInstanceInfo) {
	for _, handler := range d.handlersByEventyType[eventType] {
		handler.TriggerEvent(ctx, instance)
	}
}

func (d *BaseDiscovery) TriggerEvent(eventType ServiceEventType, instance ServiceInstanceInfo) {
	d.TriggerEventCtx(context.Background(), eventType, instance)
}

func Hostname(allowPanic ...bool) string {
	hostname, err := os.Hostname()
	if err != nil {
		if len(allowPanic) > 0 || allowPanic[0] {
			panic(err)
		}
		log.Errorf("Failed to get hostname: %s", err)
	}
	return hostname
}
