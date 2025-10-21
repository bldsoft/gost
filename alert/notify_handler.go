package alert

import (
	"context"

	_ "embed"

	"github.com/bldsoft/gost/log"
)

func NotifyHandler(notifyService *NotifyServiceAdapter) Handler {
	return HandlerFunc(func(ctx context.Context, alert ...Alert) {
		for _, alert := range alert {
			if err := notifyService.Send(ctx, alert); err != nil {
				log.FromContext(ctx).ErrorfWithFields(log.Fields{
					"alert": alert,
					"error": err,
				}, "Failed to send notification")
			}
		}
	})
}
