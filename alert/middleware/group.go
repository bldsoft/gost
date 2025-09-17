package middleware

import (
	"cmp"
	"context"
	"sync"
	"time"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/utils/errgroup"
)

type GroupConfig struct {
	GroupInterval time.Duration
	Aggregate     func([]alert.Alert, []alert.Alert) []alert.Alert
}

var DefaultGroupConfig = GroupConfig{
	GroupInterval: 5 * time.Minute,
	Aggregate:     GroupFirst,
}

func Group(ctx context.Context, config ...GroupConfig) (alert.Middleware, func()) {
	cfg := DefaultGroupConfig
	if len(config) > 0 {
		cfg = config[0]
		cfg.GroupInterval = cmp.Or(cfg.GroupInterval, DefaultGroupConfig.GroupInterval)
		if cfg.Aggregate == nil {
			cfg.Aggregate = DefaultGroupConfig.Aggregate
		}
	}

	var eg errgroup.Group
	return func(next alert.AlertHandler) alert.AlertHandler {
			var group []alert.Alert
			var groupMtx sync.Mutex

			startTimer := func() {
				eg.Go(func() error {
					select {
					case <-ctx.Done():
					case <-time.After(cfg.GroupInterval):
					}
					groupMtx.Lock()
					var current []alert.Alert
					group, current = nil, group
					groupMtx.Unlock()

					if len(current) > 0 {
						next.Handle(ctx, current...)
					}
					return nil
				})

			}

			return alert.AlertHandlerFunc(func(ctx context.Context, alerts ...alert.Alert) {
				if len(alerts) == 0 {
					return
				}
				groupMtx.Lock()
				defer groupMtx.Unlock()
				newGroup := len(group) == 0
				group = cfg.Aggregate(group, alerts)
				if newGroup && len(group) > 0 {
					startTimer()
				}
			})
		}, func() {
			eg.Wait()
		}
}

func GroupFirst(groupped []alert.Alert, newAlerts []alert.Alert) []alert.Alert {
	if len(groupped) != 0 || len(newAlerts) == 0 {
		return groupped
	}
	return newAlerts[:1]
}
