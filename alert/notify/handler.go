package notify

import (
	"context"
	"slices"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils/seq"
)

type Handler struct {
	commonNotifiers []alert.Notifier
}

func NewHandler(commonNotifiers ...alert.Notifier) *Handler {
	return &Handler{commonNotifiers: commonNotifiers}
}

func (h *Handler) Handle(ctx context.Context, alerts ...alert.Alert) {
	for _, alert := range alerts {
		for _, notifier := range seq.Concat2(
			slices.All(alert.Notifiers),
			slices.All(h.commonNotifiers),
		) {
			err := notifier.Send(ctx, alert)
			if err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{
					"alert":    alert,
					"notifier": notifier,
				}, "Failed to send alert: %v", err)
			}
		}
	}
}
