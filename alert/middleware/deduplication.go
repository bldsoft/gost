package middleware

import (
	"context"
	"time"

	"github.com/bldsoft/gost/alert"
	"github.com/jellydator/ttlcache/v3"
)

func Deduplication(ttl time.Duration) alert.Middleware {
	cache := ttlcache.New(
		ttlcache.WithTTL[string, alert.Alert](ttl),
		ttlcache.WithDisableTouchOnHit[string, alert.Alert](),
	)

	return func(next alert.AlertHandler) alert.AlertHandler {
		return alert.AlertHandlerFunc(func(ctx context.Context, alerts ...alert.Alert) {
			for _, a := range alerts {
				_, exists := cache.GetOrSet(a.ID, a)
				if exists {
					continue
				}
				cache.DeleteExpired()
				next.Handle(ctx, a)
			}
		})
	}
}
