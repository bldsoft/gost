package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/utils"
)

func Deduplication(duplicatedCache cache.Repository[*alert.Alert], uniqueKey ...func(alert alert.Alert) string) alert.Middleware {
	if len(uniqueKey) == 0 {
		uniqueKey = []func(alert alert.Alert) string{
			func(alert alert.Alert) string {
				if alert.To.IsZero() {
					return fmt.Sprintf("%s-%s-%d", alert.SourceID, alert.Severity, alert.From.Unix())
				}
				return fmt.Sprintf("%s-%s-%d", alert.SourceID, alert.Severity, alert.To.Unix())
			},
		}
	}

	return func(next alert.Handler) alert.Handler {
		return alert.HandlerFunc(func(ctx context.Context, alerts ...alert.Alert) {
			logger := log.FromContext(ctx).WithFields(log.Fields{"component": "alerts deduplication"})
			deduplicated := make([]alert.Alert, 0, len(alerts))
			for _, alert := range alerts {
				cacheKey := uniqueKey[0](alert)
				exists, err := duplicatedCache.Get(cacheKey)
				if err != nil && !errors.Is(err, utils.ErrObjectNotFound) {
					logger.ErrorWithFields(log.Fields{"err": err}, "failed to check if alert already exists in cache")
					continue
				}
				if exists != nil {
					continue
				}

				if err = duplicatedCache.Set(cacheKey, &alert); err != nil {
					logger.ErrorWithFields(log.Fields{"err": err}, "failed to set alert in cache")
					continue
				}
				deduplicated = append(deduplicated, alert)
			}
			next.Handle(ctx, deduplicated...)
		})
	}
}
