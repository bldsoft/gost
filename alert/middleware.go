package alert

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/log"
)

const (
	deduplicationWaitingPrefix  = "dw"
	deduplicationReceivedPrefix = "dr"
	groupingStoredPrefix        = "gs"
	groupingIgnoredStartPrefix  = "gi"
)

func GroupRecurringMiddleware(store cache.Repository[Alert], period time.Duration) Middleware {
	return func(next Handler) Handler {
		return AlertHandlerFunc(func(ctx context.Context, alerts ...Alert) {
			grouped := make([]Alert, 0, len(alerts))

			for _, alert := range alerts {
				key := fmt.Sprintf("%s-%s-%d", groupingStoredPrefix, alert.SourceID, alert.Severity)
				ignoreKey := fmt.Sprintf("%s-%s-%d", groupingIgnoredStartPrefix, alert.SourceID, alert.Severity)

				storedAlert, err := store.Get(key)
				if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
					log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to get value from store")
				}

				// we received start/end/happened alert outside ignore period
				isOutsideIgnorePeriod := err != nil || !storedAlert.To.Add(period).After(alert.From)

				if !isOutsideIgnorePeriod {
					// we received alert during ignore period (either start/end/happened)
					if shouldSkipAlert(ctx, store, alert, storedAlert, ignoreKey, period) {
						continue
					}
				}

				// process alert outside ignore period or happened alert that started in ignore period but finished outside
				processAlertOutsideIgnorePeriod(ctx, store, &alert, ignoreKey, key, period)
				grouped = append(grouped, alert)
			}

			next.Handle(ctx, grouped...)
		})
	}
}

func shouldSkipAlert(ctx context.Context, store cache.Repository[Alert], alert Alert, storedAlert Alert, ignoreKey string, period time.Duration) bool {
	// this alert started during ignore period
	if alert.To.IsZero() {
		// we store it in case it finishes outside ignore period
		if err := store.Set(ignoreKey, alert); err != nil {
			log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to store ignored start alert")
		}
		return true
	}

	// this alert started during ignore period and finished during ignore period, we skip it
	if alert.To.Before(storedAlert.To.Add(period)) {
		if err := store.Delete(ignoreKey); err != nil {
			log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to delete ignored start alert")
		}
		return true
	}

	// this is happened alert which started during ignore period and finished outside ignore period
	return false
}

func processAlertOutsideIgnorePeriod(ctx context.Context, store cache.Repository[Alert], alert *Alert, ignoreKey, key string, period time.Duration) {
	// this is end or happened alert
	if !alert.To.IsZero() {
		// check if we had start alert during ignore period
		ignoredStartAlert, err := store.Get(ignoreKey)
		if err == nil {
			// this is end alert
			alert.From = ignoredStartAlert.From
			if err := store.Delete(ignoreKey); err != nil {
				log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to delete ignored start alert")
			}
		} else if !errors.Is(err, cache.ErrCacheMiss) {
			log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to get ignored start alert from store")
		}
	}

	// just in case use ttl
	if err := store.SetFor(key, *alert, 2*period); err != nil {
		log.FromContext(ctx).ErrorWithFields(log.Fields{"err": err}, "alerts recurring middleware: failed to store alert")
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
				key := fmt.Sprintf("%s-%s-%d", deduplicationWaitingPrefix, alert.SourceID, alert.Severity)

				//for alerts that were already finished (in case of sender restarts etc.)
				timedKey := fmt.Sprintf("%s-%s-%d", deduplicationReceivedPrefix, alert.SourceID, alert.Severity)
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
				// is "happened" (started and finished in the same window) alert
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
