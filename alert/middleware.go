package alert

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/log"
)

func GroupRecurringMiddleware(store cache.Repository[time.Time], period time.Duration) Middleware {
	return func(next Handler) Handler {
		return AlertHandlerFunc(func(ctx context.Context, alerts ...Alert) {
			var grouped []Alert

			for _, alert := range alerts {
				to, err := store.Get(alert.SourceID)
				if err == nil {
					if to.After(time.Now()) {
						continue
					}
				} else if !errors.Is(err, cache.ErrCacheMiss) {
					log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts reccuring middleware: failed to get value from store")
				}

				if store.SetFor(alert.SourceID, time.Now().Add(period), period) != nil {
					log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts reccuring middleware: failed to store alert")
				}

				grouped = append(grouped, alert)
			}

			next.Handle(ctx, grouped...)
		})
	}
}

// TODO: handle restart cases (when we get previous alerts that were completely sent)
func DeduplicationMiddleware(store cache.Repository[Alert], ttl time.Duration) Middleware {
	return func(next Handler) Handler {
		return AlertHandlerFunc(func(ctx context.Context, alerts ...Alert) {
			var deduplicated []Alert
			for _, alert := range alerts {
				key := fmt.Sprintf("%s-%d", alert.SourceID, alert.Severity)

				_, err := store.Get(key)

				if alert.To.IsZero() {
					if err == nil {
						continue //if we have this alert in store it means that we didnt receive finishing for this alert
					} else {
						if store.SetFor(key, alert, ttl) != nil {
							fmt.Println("failed to set cache alert with fast key ", key, " err ", err)
						}
					}
				} else {
					if err == nil {
						if store.Delete(key) != nil {
							fmt.Println("failed to delete cache alert with fast key ", key, " err ", err)
						} //we received finishing for this alert and can delete it from store
					} else {
						continue //weve already received finisher for this alert
					}
				}

				deduplicated = append(deduplicated, alert)
			}

			next.Handle(ctx, deduplicated...)
		})
	}
}
