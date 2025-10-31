package middleware

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
)

// use after deduplication middleware only
func LogicalDeduplication(duplicatedCache cache.Repository[*alert.Alert], deduplicateStartDur time.Duration, alertTypeKey ...func(alert alert.Alert) string) alert.Middleware {
	if len(alertTypeKey) == 0 {
		alertTypeKey = []func(alert alert.Alert) string{
			func(alert alert.Alert) string {
				return fmt.Sprintf("%s-%s", alert.SourceID, alert.Severity)
			},
		}
	}
	return func(next alert.Handler) alert.Handler {
		return alert.HandlerFunc(func(ctx context.Context, alerts ...alert.Alert) {
			logger := log.FromContext(ctx).WithFields(log.Fields{"component": "alerts logical deduplication"})
			deduplicated := make([]alert.Alert, 0, len(alerts))

			pass := func(cacheKey string, alert alert.Alert) {
				if err := duplicatedCache.Set(cacheKey, &alert); err != nil {
					logger.ErrorWithFields(log.Fields{"err": err}, "failed to set alert in cache")
					return
				}
				deduplicated = append(deduplicated, alert)
			}

			for _, alert := range alerts {
				cacheKey := alertTypeKey[0](alert)
				prev, err := duplicatedCache.Get(cacheKey)
				if err != nil && !errors.Is(err, utils.ErrObjectNotFound) {
					logger.ErrorWithFields(log.Fields{"err": err}, "failed to check if alert already exists in cache")
					continue
				}

				if !alert.To.IsZero() {
					if prev != nil && prev.To.IsZero() {
						alert.From = prev.From
					}
					pass(cacheKey, alert)
					continue
				}

				if prev != nil && prev.To.IsZero() && alert.From.Sub(prev.From) < deduplicateStartDur {
					continue
				}
				pass(cacheKey, alert)
			}
			if len(deduplicated) == 0 {
				return
			}
			next.Handle(ctx, deduplicated...)
		})
	}
}
