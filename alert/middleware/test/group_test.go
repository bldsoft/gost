package test

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/alert/middleware"
	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/cache/bigcache"
	"github.com/stretchr/testify/require"
)

func TestGroupMiddleware(t *testing.T) {
	alertLow1 := AlertFactory(time.Time{}, "1", alert.SeverityLevelLow)
	testcases := []*groupTestcaseBuilder{
		newGroupTestcaseBuilder("single active alert").
			Interval(5 * time.Minute).GroupInterval(15 * time.Minute).
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts(alertLow1(0)).ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			ExpectedGroupedAlerts(),
		newGroupTestcaseBuilder("single finished alert").
			Interval(5 * time.Minute).GroupInterval(25 * time.Minute).
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts(alertLow1(0)).ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0, 20*time.Minute)).ExpectedImmediateAlerts(alertLow1(0, 20*time.Minute)).ExpectedGroupedAlerts().
			ExpectedGroupedAlerts(),
		newGroupTestcaseBuilder("not grouped alert").
			Interval(5 * time.Minute).GroupInterval(10 * time.Minute).
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts(alertLow1(0)).ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0, 5*time.Minute)).ExpectedImmediateAlerts(alertLow1(0, 5*time.Minute)).ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(10 * time.Minute)).ExpectedImmediateAlerts(alertLow1(10 * time.Minute)).ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(10*time.Minute, 20*time.Minute)).ExpectedImmediateAlerts(alertLow1(10*time.Minute, 20*time.Minute)).ExpectedGroupedAlerts().
			EmitAlerts().ExpectedImmediateAlerts().ExpectedGroupedAlerts(),
		newGroupTestcaseBuilder("grouped alert").
			Interval(5 * time.Minute).GroupInterval(25 * time.Minute).
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts(alertLow1(0)).ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0, 5*time.Minute)).ExpectedImmediateAlerts(alertLow1(0, 5*time.Minute)).ExpectedGroupedAlerts().
			EmitAlerts().ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(15 * time.Minute)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(15*time.Minute, 25*time.Minute)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			ExpectedGroupedAlerts(alertLow1(0, 25*time.Minute).AddMetaData("count", "2")),
		newGroupTestcaseBuilder("suppressed start, not suppressed end").
			Interval(5*time.Minute).GroupInterval(15*time.Minute).
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts(alertLow1(0)).ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0, 10*time.Minute), alertLow1(11*time.Minute)).ExpectedImmediateAlerts(alertLow1(0, 10*time.Minute)).ExpectedGroupedAlerts().
			EmitAlerts().ExpectedImmediateAlerts().ExpectedGroupedAlerts(alertLow1(0).AddMetaData("count", "2")).
			EmitAlerts(alertLow1(11*time.Minute, 25*time.Minute)).ExpectedImmediateAlerts(alertLow1(11*time.Minute, 25*time.Minute)).ExpectedGroupedAlerts().
			EmitAlerts().ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(31*time.Minute), alertLow1(31*time.Minute, 35*time.Minute)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			ExpectedImmediateAlerts().ExpectedGroupedAlerts(alertLow1(11*time.Minute, 35*time.Minute).AddMetaData("count", "2")),
		newGroupTestcaseBuilder("suppressed and shifted start, not suppressed end").
			Interval(5*time.Minute).GroupInterval(15*time.Minute).
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts(alertLow1(0)).ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(0, 10*time.Minute), alertLow1(11*time.Minute)).ExpectedImmediateAlerts(alertLow1(0, 10*time.Minute)).ExpectedGroupedAlerts().
			EmitAlerts().ExpectedImmediateAlerts().ExpectedGroupedAlerts(alertLow1(0).AddMetaData("count", "2")).
			EmitAlerts(alertLow1(21*time.Minute, 25*time.Minute)).ExpectedImmediateAlerts(alertLow1(11*time.Minute, 25*time.Minute)).ExpectedGroupedAlerts().
			EmitAlerts().ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			EmitAlerts(alertLow1(31*time.Minute), alertLow1(31*time.Minute, 35*time.Minute)).ExpectedImmediateAlerts().ExpectedGroupedAlerts().
			ExpectedImmediateAlerts().ExpectedGroupedAlerts(alertLow1(11*time.Minute, 35*time.Minute).AddMetaData("count", "2")),
	}
	for _, testcase := range testcases {
		testcase.Run(t)
	}
}

func TestGroupMiddlewareClose(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		alertLow1 := AlertFactory(time.Time{}, "1", alert.SeverityLevelLow)
		store := cache.Typed[*alert.Alert](bigcache.NewExpiringRepository("{}"))
		groupRep := newTestGroupRepository()
		group, close := middleware.NewGroupMiddleware("1", groupRep, 10*time.Minute).Middleware()
		defer close()
		middleware := alert.Middlewares(
			middleware.Sort,
			middleware.Deduplication(store),
			middleware.LogicalDeduplication(store, 1*time.Hour),
			group,
		)
		recorder := &mockHandler{}
		handler := middleware(recorder)
		handler.Handle(context.Background(), alertLow1(0), alertLow1(0, 5*time.Minute), alertLow1(6*time.Minute), alertLow1(6*time.Minute, 7*time.Minute))

		require.Equal(t, 2, len(recorder.ReceivedAlerts()))
		recorder.Reset()
		close()
		require.Equal(t, 1, len(recorder.ReceivedAlerts()))
		require.True(t, recorder.ReceivedAlerts()[0].MetaData["count"] == "2")
	})
}

type groupTestcaseBuilder struct {
	name          string
	interval      time.Duration
	groupInterval time.Duration

	emittedAlerts           [][]alert.Alert
	expectedImmediateAlerts [][]alert.Alert
	expectedGroupedAlerts   [][]alert.Alert
}

func newGroupTestcaseBuilder(name string) *groupTestcaseBuilder {
	return new(groupTestcaseBuilder).
		Name(name).
		GroupInterval(10 * time.Minute).
		Interval(10 * time.Minute)
}

func (b *groupTestcaseBuilder) Name(name string) *groupTestcaseBuilder {
	b.name = name
	return b
}

func (b *groupTestcaseBuilder) Interval(interval time.Duration) *groupTestcaseBuilder {
	b.interval = interval
	return b
}

func (b *groupTestcaseBuilder) GroupInterval(interval time.Duration) *groupTestcaseBuilder {
	b.groupInterval = interval
	return b
}

func (b *groupTestcaseBuilder) EmitAlerts(alerts ...alert.Alert) *groupTestcaseBuilder {
	b.emittedAlerts = append(b.emittedAlerts, alerts)
	return b
}

func (b *groupTestcaseBuilder) ExpectedImmediateAlerts(alerts ...alert.Alert) *groupTestcaseBuilder {
	b.expectedImmediateAlerts = append(b.expectedImmediateAlerts, alerts)
	return b
}

func (b *groupTestcaseBuilder) ExpectedGroupedAlerts(alerts ...alert.Alert) *groupTestcaseBuilder {
	b.expectedGroupedAlerts = append(b.expectedGroupedAlerts, alerts)
	return b
}

func (b *groupTestcaseBuilder) adjustAlerts(now time.Time, alerts []alert.Alert) {
	for i := range alerts {
		alerts[i].From = now.Add(alerts[i].From.Sub(time.Time{}))
		if !alerts[i].To.IsZero() {
			alerts[i].To = now.Add(alerts[i].To.Sub(time.Time{}))
		}
	}
}

func (b *groupTestcaseBuilder) Run(t *testing.T) {
	t.Run(b.name, func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			store := cache.Typed[*alert.Alert](bigcache.NewExpiringRepository("{}"))
			groupRep := newTestGroupRepository()
			checkExpiredGroupPeriod := 1 * time.Minute
			group, close := middleware.NewGroupMiddleware("1", groupRep, b.groupInterval).
				WithCheckExpiredGroupPeriod(checkExpiredGroupPeriod).Middleware()
			defer close()
			middleware := alert.Middlewares(
				middleware.Sort,
				middleware.Deduplication(store),
				middleware.LogicalDeduplication(store, 1*time.Hour),
				group,
			)
			recorder := &mockHandler{}
			handler := middleware(recorder)

			now := time.Now()
			intervalCaller := IntervalCaller(now, b.interval)
			emittedAlerts := make([]HandleCall, 0, len(b.emittedAlerts))
			for _, alerts := range b.emittedAlerts {
				b.adjustAlerts(now, alerts)
				emittedAlerts = append(emittedAlerts, intervalCaller(alerts...))
			}

			expectedImmediateAlerts := make([][]alert.Alert, 0, len(b.expectedImmediateAlerts))
			for _, alerts := range b.expectedImmediateAlerts {
				b.adjustAlerts(now, alerts)
				expectedImmediateAlerts = append(expectedImmediateAlerts, alerts)
			}
			expectedGroupedAlerts := make([][]alert.Alert, 0, len(b.expectedGroupedAlerts))
			for _, alerts := range b.expectedGroupedAlerts {
				b.adjustAlerts(now, alerts)
				expectedGroupedAlerts = append(expectedGroupedAlerts, alerts)
			}

			go run(t, handler, emittedAlerts...)
			synctest.Wait()
			time.Sleep(checkExpiredGroupPeriod)
			for i := 0; i < max(len(expectedImmediateAlerts), len(expectedGroupedAlerts)); i++ {
				time.Sleep(b.interval)

				isGrouped := func(a alert.Alert) bool {
					_, grouped := a.MetaData["count"]
					return grouped
				}

				t.Logf("checking immediate alerts %d", i)
				immediateAlerts := recorder.RemoveFunc(func(a alert.Alert) bool {
					return !isGrouped(a)
				})
				if i < len(expectedImmediateAlerts) {
					RequireEqualAlerts(t, expectedImmediateAlerts[i], immediateAlerts, now)
				} else {
					RequireEqualAlerts(t, []alert.Alert{}, immediateAlerts, now)
				}

				t.Logf("checking grouped alerts %d", i)
				groupedAlerts := recorder.RemoveFunc(isGrouped)
				if i < len(expectedGroupedAlerts) {
					RequireEqualAlerts(t, expectedGroupedAlerts[i], groupedAlerts, now)
				} else {
					RequireEqualAlerts(t, []alert.Alert{}, groupedAlerts, now)
				}
			}
		})
	})
}
