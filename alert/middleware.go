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
			grouped := make([]Alert, 0, len(alerts))

			for _, alert := range alerts {
				key := fmt.Sprintf("group-%s-%d", alert.SourceID, alert.Severity)
				to, err := store.Get(key)
				if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
					log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to get value from store")
				}
				now := time.Now()

				// Skip if alert is still within cooldown period
				if err == nil && to.After(now) {
					continue
				}

				if err := store.SetFor(key, now.Add(period), period); err != nil {
					log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to store alert")
				}

				grouped = append(grouped, alert)
			}

			next.Handle(ctx, grouped...)
		})
	}
}

// DeduplicationMiddleware TTL is being used for caching alerts to prevent duplicates on sender restarts etc
func DeduplicationMiddleware(store cache.Repository[Alert], TTL time.Duration) Middleware {
	return func(next Handler) Handler {
		return AlertHandlerFunc(func(ctx context.Context, alerts ...Alert) {
			deduplicated := make([]Alert, 0, len(alerts))
			logger := log.FromContext(ctx)

			for _, alert := range alerts {
				//for alerts where we are waiting for the finisher
				key := fmt.Sprintf("dedup-%s-%d-wait", alert.SourceID, alert.Severity)

				//for alerts that were already finished (in case of sender restarts etc.)
				timedKey := fmt.Sprintf("dedup-%s-%d-recvd", alert.SourceID, alert.Severity)
				if val, err := store.Get(timedKey); err == nil {
					if val.To.After(alert.From) {
						continue //already received this alert
					}
				}

				_, err := store.Get(key)
				alertExistsInStore := err == nil
				isStartAlert := alert.To.IsZero()

				if isStartAlert {
					if alertExistsInStore {
						continue // already received this start alert
					}
					if err := store.Set(key, alert); err != nil {
						logger.ErrorWithFields(log.Fields{"err": err, "key": key}, "alerts deduplication middleware: failed to store alert in cache with normal key")
					}
					deduplicated = append(deduplicated, alert)
					continue
				}

				//either delete entry if this is an end alert or fail if this
				// is "happened" (started in finished in the same window) alert
				// that we didn't store, we don't care
				_ = store.Delete(key)

				// Store in case we get this alert again (e.g. sender restarts)
				if err := store.SetFor(timedKey, alert, TTL); err != nil {
					logger.ErrorWithFields(log.Fields{"err": err, "key": timedKey}, "alerts deduplication middleware: failed to store received alert in cache with timed key")
				}

				deduplicated = append(deduplicated, alert)
			}

			next.Handle(ctx, deduplicated...)
		})
	}
}
