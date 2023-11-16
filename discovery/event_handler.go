package discovery

import (
	"context"
	"sync"
)

//go:generate go run github.com/abice/go-enum -f=$GOFILE

// ENUM(discovered, up, down, removed, count)
type EventType int

type EventHandler struct {
	handler func(ctx context.Context, instance ServiceInstanceInfo)

	once       *sync.Once
	eventTypes []EventType
	filters    []func(ctx context.Context, instance ServiceInstanceInfo) bool
}

func NewEventHandler(handler func(ctx context.Context, instance ServiceInstanceInfo)) *EventHandler {
	return &EventHandler{handler: handler}
}

func (h *EventHandler) ServiceName(name string) *EventHandler {
	h.filters = append(h.filters, func(ctx context.Context, instance ServiceInstanceInfo) bool {
		return instance.ServiceName == name
	})
	return h
}

func (h *EventHandler) EventTypes() []EventType {
	return h.eventTypes
}

// setter
func (h *EventHandler) EventType(eventType EventType, eventTypes ...EventType) *EventHandler {
	h.eventTypes = append(h.eventTypes, eventType)
	h.eventTypes = append(h.eventTypes, eventTypes...)
	return h
}

// setter
func (h *EventHandler) Once() *EventHandler {
	h.once = new(sync.Once)
	return h
}

func (h *EventHandler) SetHandler(handler func(ctx context.Context, instance ServiceInstanceInfo)) *EventHandler {
	h.handler = handler
	return h
}

func (h *EventHandler) Node(node string) *EventHandler {
	h.filters = append(h.filters, func(ctx context.Context, instance ServiceInstanceInfo) bool {
		return instance.Node == node
	})
	return h
}

func (h *EventHandler) TriggerEvent(ctx context.Context, instance ServiceInstanceInfo) {
	for _, f := range h.filters {
		if !f(ctx, instance) {
			return
		}
	}
	if h == nil {
		return
	}

	if h.once != nil {
		h.once.Do(func() { h.handler(ctx, instance) })
	} else {
		h.handler(ctx, instance)
	}
}
