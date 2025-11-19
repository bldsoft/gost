package test

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/alert/middleware"
	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/cache/bigcache"
	"github.com/bldsoft/gost/utils/seq"
	"github.com/stretchr/testify/require"
)

func TestDeduplicationMiddleware(t *testing.T) {
	now := time.Now()
	alertLow1 := AlertFactory(now, "1", alert.SeverityLevelLow)
	testcases := []struct {
		name             string
		handlerCalls     []HandleCall
		expectSentAlerts []alert.Alert
	}{
		{
			name: "recurring alerts",
			handlerCalls: func() []HandleCall {
				source1 := IntervalCaller(now, 5*time.Minute)
				return []HandleCall{
					source1(alertLow1(0)),
					source1(alertLow1(0, 5*time.Minute), alertLow1(6*time.Minute)),
					source1(alertLow1(6 * time.Minute)),
				}
			}(),
			expectSentAlerts: []alert.Alert{
				alertLow1(0),
				alertLow1(0, 5*time.Minute),
				alertLow1(6 * time.Minute),
			},
		},
		{
			name: "duplicated alert start",
			handlerCalls: func() []HandleCall {
				source1 := IntervalCaller(now, 5*time.Minute)
				return []HandleCall{
					source1(alertLow1(0)),
					source1(alertLow1(0)),
					source1(alertLow1(0), alertLow1(0)),
				}
			}(),
			expectSentAlerts: []alert.Alert{alertLow1(0)},
		},
		{
			name: "duplicated alert end",
			handlerCalls: func() []HandleCall {
				source1 := IntervalCaller(now, 5*time.Minute)
				return []HandleCall{
					source1(alertLow1(0)),
					source1(alertLow1(0)),
					source1(alertLow1(0, 11*time.Minute), alertLow1(0, 11*time.Minute)), // end
					source1(alertLow1(0, 11*time.Minute)),
				}
			}(),
			expectSentAlerts: []alert.Alert{alertLow1(0), alertLow1(0, 11*time.Minute)},
		}, {
			name: "start shift",
			handlerCalls: func() []HandleCall {
				source1 := IntervalCaller(now, 5*time.Minute)
				return []HandleCall{
					source1(alertLow1(0)),
					source1(alertLow1(5 * time.Minute)),
					source1(alertLow1(10 * time.Minute)),
					source1(alertLow1(15*time.Minute, 20*time.Minute)),
					source1(alertLow1(20*time.Minute, 20*time.Minute)),
				}
			}(),
			expectSentAlerts: []alert.Alert{alertLow1(0), alertLow1(0, 20*time.Minute)},
		}, {
			name: "finished alert shift",
			handlerCalls: func() []HandleCall {
				source1 := IntervalCaller(now, 5*time.Minute)
				return []HandleCall{
					source1(alertLow1(0)),
					source1(alertLow1(0*time.Minute, 5*time.Minute)),
					source1(alertLow1(2*time.Minute, 7*time.Minute)),
				}
			}(),
			expectSentAlerts: []alert.Alert{alertLow1(0), alertLow1(0, 5*time.Minute)},
		}, {
			name: "end negative shift",
			handlerCalls: func() []HandleCall {
				source1 := IntervalCaller(now, 5*time.Minute)
				return []HandleCall{
					source1(alertLow1(0)),
					source1(alertLow1(0*time.Minute, 5*time.Minute)),
					source1(alertLow1(2*time.Minute, 4*time.Minute)),
				}
			}(),
			expectSentAlerts: []alert.Alert{alertLow1(0), alertLow1(0, 5*time.Minute)},
		}, {
			name: "late alert",
			handlerCalls: func() []HandleCall {
				source1 := IntervalCaller(now, 5*time.Minute)
				return []HandleCall{
					source1(),
					source1(alertLow1(5 * time.Minute)),
					source1(alertLow1(0, 4*time.Minute)),
				}
			}(),
			expectSentAlerts: []alert.Alert{alertLow1(5 * time.Minute)},
		},
	}
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				rep := cache.Typed[*alert.Alert](bigcache.NewExpiringRepository("{}"))

				middleware := alert.Middlewares(
					middleware.Sort,
					middleware.Deduplication(rep),
					middleware.LogicalDeduplication(rep, 1*time.Hour),
				)
				recorder := &mockHandler{}
				handler := middleware(recorder)
				run(t, handler, testcase.handlerCalls...)
				RequireEqualAlerts(t, testcase.expectSentAlerts, recorder.ReceivedAlerts())
			})
		})
	}
}

func RequireEqualAlerts(t *testing.T, expected, actual []alert.Alert, now ...time.Time) {
	format := func(alerts ...alert.Alert) string {
		var sb strings.Builder
		for _, alert := range alerts {
			if len(now) > 0 {
				sb.WriteString(fmt.Sprintf("%s %s %s %s %s\n",
					alert.SourceID,
					alert.Severity,
					alert.From.Sub(now[0]),
					alert.To.Sub(now[0]),
					alert.MetaData))
			} else {
				sb.WriteString(fmt.Sprintf("%s %s %s %s %s\n",
					alert.SourceID,
					alert.Severity,
					alert.From,
					alert.To,
					alert.MetaData))
			}
		}
		return sb.String()
	}
	require.Equal(t, len(expected), len(actual),
		"expected %v alerts, got %v", format(expected...), format(actual...))
	for i, expect := range expected {
		actual := actual[i]
		require.Equal(t, expect.SourceID, actual.SourceID, "source id: %s, %s", format(expect), format(actual))
		require.Equal(t, expect.Severity, actual.Severity, "severity: %s, %s", format(expect), format(actual))
		require.WithinDuration(t, expect.From, actual.From, 5*time.Millisecond, "from: %s, %s", format(expect), format(actual))
		require.WithinDuration(t, expect.To, actual.To, 5*time.Millisecond, "to: %s, %s", format(expect), format(actual))
		require.Equal(t, expect.MetaData, actual.MetaData, "meta data: %s, %s", format(expect), format(actual))
		require.Equal(t, expect.Receivers, actual.Receivers, "receivers: %s, %s", format(expect), format(actual))
	}
}

type HandleCall struct {
	Alerts   []alert.Alert
	CallTime time.Time
}

func AlertFactory(now time.Time, id string, severity alert.SeverityLevel) func(start time.Duration, end ...time.Duration) alert.Alert {
	return func(start time.Duration, end ...time.Duration) alert.Alert {
		res := alert.Alert{
			SourceID: id,
			Severity: severity,
			From:     now.Add(start),
		}
		if len(end) > 0 {
			res.To = now.Add(end[0])
		}
		return res
	}
}

func IntervalCaller(start time.Time, window time.Duration) func(alerts ...alert.Alert) HandleCall {
	now := start
	return func(alerts ...alert.Alert) HandleCall {
		now = now.Add(window)
		res := HandleCall{
			Alerts:   alerts,
			CallTime: now,
		}
		return res
	}
}

func run(t *testing.T, handler alert.Handler, handleCalls ...HandleCall) {
	slices.SortFunc(handleCalls, func(a, b HandleCall) int {
		return a.CallTime.Compare(b.CallTime)
	})
	for _, handleCall := range handleCalls {
		time.Sleep(time.Until(handleCall.CallTime))
		handler.Handle(context.Background(), handleCall.Alerts...)
	}
}

type mockHandler struct {
	receivedAlerts []alert.Alert
	mtx            sync.RWMutex
}

func (m *mockHandler) Handle(_ context.Context, alerts ...alert.Alert) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.receivedAlerts = append(m.receivedAlerts, alerts...)
}

func (m *mockHandler) ReceivedAlerts() []alert.Alert {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return slices.Clone(m.receivedAlerts)
}

func (m *mockHandler) RemoveFunc(remove func(alert.Alert) bool) (deleted []alert.Alert) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	res := slices.Collect(seq.FilterFunc(slices.Values(m.receivedAlerts), remove))
	m.receivedAlerts = slices.DeleteFunc(m.receivedAlerts, remove)
	return res
}

func (m *mockHandler) Reset() {
	m.receivedAlerts = nil
}
