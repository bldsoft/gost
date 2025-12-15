package middleware

import (
	"context"

	"github.com/bldsoft/gost/alert"
)

func FillMetadata(metadata map[string]any) alert.Middleware {
	return func(next alert.Handler) alert.Handler {
		return alert.HandlerFunc(func(ctx context.Context, alerts ...alert.Alert) {
			for i := range alerts {
				for k, v := range metadata {
					alerts[i] = alerts[i].AddMetaData(k, v)
				}
			}
			next.Handle(ctx, alerts...)
		})
	}
}
