package test

import (
	"context"
	"slices"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bldsoft/gost/alert"
)

type AlertGen struct {
	gen      func(i int) alert.Alert
	count    int
	interval time.Duration

	alerts atomic.Pointer[[]alert.Alert]
}

func NewAlertGen(count int, interval time.Duration) *AlertGen {
	res := &AlertGen{
		gen: func(i int) alert.Alert {
			return alert.Alert{
				ID:       strconv.Itoa(i),
				Severity: alert.SeverityLow,
				From:     time.Now(),
			}
		},
		count:    count,
		interval: interval,
	}
	res.alerts.Store(new([]alert.Alert))
	return res
}

func (a *AlertGen) Start(ctx context.Context) {
	tick := time.Tick(a.interval)
	for i := range a.count {
		alerts := a.alerts.Load()
		res := slices.Clone(*alerts)
		res = append(res, a.gen(i))
		a.alerts.Store(&res)
		select {
		case <-ctx.Done():
			return
		case <-tick:
		}
	}
}

func (a *AlertGen) CheckAlerts(ctx context.Context) ([]alert.Alert, error) {
	return *a.alerts.Load(), nil
}
