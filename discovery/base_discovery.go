package discovery

import "context"

type BaseDiscovery struct {
	handlersByEventyType [ServiceEventTypeCount][]EventServiceHandler
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

func (d *BaseDiscovery) TriggerEventCtx(ctx context.Context, eventType ServiceEventType, instance *ServiceInstanceInfoFull) {
	for _, handler := range d.handlersByEventyType[eventType] {
		handler.TriggerEvent(ctx, instance)
	}
}

func (d *BaseDiscovery) TriggerEvent(eventType ServiceEventType, instance *ServiceInstanceInfoFull) {
	d.TriggerEventCtx(context.Background(), eventType, instance)
}
