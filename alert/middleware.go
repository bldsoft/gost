package alert

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bldsoft/gost/cache"
)

func FilterBySeverityMiddleware(minSeverity SeverityLevel) Middleware {
	return func(next Handler) Handler {
		return AlertHandlerFunc(func(ctx context.Context, alerts ...Alert) {
			var filtered []Alert
			for _, alert := range alerts {
				if alert.Severity >= minSeverity {
					filtered = append(filtered, alert)
				}
			}
			if len(filtered) > 0 {
				next.Handle(ctx, filtered...)
			}
		})
	}
}

func DeduplicationMiddleware(store cache.Repository[Alert], ttl time.Duration) Middleware {
	return func(next Handler) Handler {
		return AlertHandlerFunc(func(ctx context.Context, alerts ...Alert) {
			var deduplicated []Alert
			for _, alert := range alerts {
				uniqueKey := fmt.Sprintf("%s:%d:%d", alert.SourceID, alert.From.Unix(), alert.To.Unix())
				if _, err := store.Get(uniqueKey); err != cache.ErrCacheMiss { //TODO: are we always hoping for cache miss?
					continue
				}

				deduplicated = append(deduplicated, alert)
				err := store.SetFor(uniqueKey, alert, ttl)
				if err != nil {
					log.Printf("Failed to set alert in store: %v", err)
				}
			}

			if len(deduplicated) == 0 {
				return
			}

			next.Handle(ctx, deduplicated...)
		})
	}
}
