package middleware

import (
	"context"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/log"
)

type AlertLog interface {
	UpsertMany(ctx context.Context, alerts ...alert.Alert) error
}

func Log(alertLog AlertLog) alert.Middleware {
	return func(next alert.Handler) alert.Handler {
		return alert.HandlerFunc(func(ctx context.Context, alerts ...alert.Alert) {
			if len(alerts) == 0 {
				return
			}
			if err := alertLog.UpsertMany(ctx, alerts...); err != nil {
				log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "failed to insert alerts into log")
				return
			}
			next.Handle(ctx, alerts...)
		})
	}
}
