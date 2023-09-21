package discovery

import "context"

//go:generate go run github.com/abice/go-enum -f=$GOFILE

// ENUM(discovered, up, down, removed, count)
type ServiceEventType int

type EventServiceHandler struct {
	eventTypes []ServiceEventType
	handler    func(ctx context.Context, instance ServiceInstanceInfo)
	filters    []func(ctx context.Context, instance ServiceInstanceInfo) bool
}

func NewEventServiceHandler(eventType ServiceEventType, eventTypes ...ServiceEventType) *EventServiceHandler {
	res := &EventServiceHandler{}
	res.eventTypes = append(res.eventTypes, eventType)
	res.eventTypes = append(res.eventTypes, eventTypes...)
	return res
}

func (h *EventServiceHandler) Types() []ServiceEventType {
	return h.eventTypes
}

func (h *EventServiceHandler) ServiceName(name string) *EventServiceHandler {
	h.filters = append(h.filters, func(ctx context.Context, instance ServiceInstanceInfo) bool {
		return instance.ServiceName == name
	})
	return h
}

func (h *EventServiceHandler) Node(node string) *EventServiceHandler {
	h.filters = append(h.filters, func(ctx context.Context, instance ServiceInstanceInfo) bool {
		return instance.Node == node
	})
	return h
}

func (h *EventServiceHandler) TriggerEvent(ctx context.Context, instance ServiceInstanceInfo) {
	for _, f := range h.filters {
		if !f(ctx, instance) {
			return
		}
	}
	h.handler(ctx, instance)
}
