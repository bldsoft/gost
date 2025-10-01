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
				to, err := store.Get(alert.SourceID)
				if err == nil {
					if to.After(time.Now()) {
						continue
					}
				} else if !errors.Is(err, cache.ErrCacheMiss) {
					log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to get value from store")
				}

				if err := store.SetFor(alert.SourceID, time.Now().Add(period), period); err != nil {
					log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to store alert")
				}

				grouped = append(grouped, alert)
			}

			next.Handle(ctx, grouped...)
		})
	}
}

// TODO: handle restart cases (when we get previous alerts that were completely sent)
func DeduplicationMiddleware(store cache.Repository[Alert], TTL time.Duration) Middleware {
	return func(next Handler) Handler {
		return AlertHandlerFunc(func(ctx context.Context, alerts ...Alert) {
			deduplicated := make([]Alert, 0, len(alerts))
			logger := log.FromContext(ctx)

			for _, alert := range alerts {
				//for alerts where we are waiting for the finisher
				key := fmt.Sprintf("%s-%d", alert.SourceID, alert.Severity)

				//for alerts that were already finished (in case of sender restarts etc)
				timedKey := fmt.Sprintf("%s-%d-recvd", alert.SourceID, alert.Severity)
				if val, err := store.Get(timedKey); err == nil {
					if val.To.After(alert.From) {
						continue //already received this alert fully previously
					}
				}

				_, err := store.Get(key)
				alertExistsInStore := err == nil
				isStartAlert := alert.To.IsZero()

				if isStartAlert {
					// Start alert: skip if already in store (duplicate), otherwise store it
					if alertExistsInStore {
						continue // Already received this start alert
					}

					if err := store.Set(key, alert); err != nil {
						logger.ErrorWithFields(log.Fields{"err": err, "key": key}, "alerts deduplication middleware: failed to store alert in cache")
					}
				} else {
					// End alert: remove from store if exists, skip if already processed
					if !alertExistsInStore {
						continue // Already received the end alert for this
					}

					if err := store.Delete(key); err != nil {
						logger.ErrorWithFields(log.Fields{"err": err, "key": key}, "alerts deduplication middleware: failed to delete alert from cache")
					}

					// Store in case we get this alert again (e.g. sender restarts)
					if err := store.SetFor(timedKey, alert, TTL); err != nil {
						logger.ErrorWithFields(log.Fields{"err": err, "key": timedKey}, "alerts deduplication middleware: failed to store received end alert in cache")
					}
				}

				deduplicated = append(deduplicated, alert)
			}

			next.Handle(ctx, deduplicated...)
		})
	}
}
