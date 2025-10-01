package alert

import (
	"context"
	"fmt"
	"maps"
	"sort"
	"testing"
	"testing/synctest"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/cache/bigcache"
	"github.com/stretchr/testify/assert"
)

const (
	UP   = true
	DOWN = false
)

func NewMockLocalCacheRepository() *cache.ExpiringCacheRepository {
	return bigcache.NewExpiringRepository("{}")
}

type mockAlert struct {
	*Alert
	sendAt time.Time
}

func (m Alert) SeverityString() string {
	switch m.Severity {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

func (m Alert) String() string {
	base := fmt.Sprintf("%s, %s: %d", m.SourceID, m.SeverityString(), m.From.Unix())
	if !m.To.IsZero() {
		base = fmt.Sprintf("%s-%d", base, m.To.Unix())
	}

	return fmt.Sprintf("%s: %s", base, m.MetaData["message"])
}

func cloneMeta(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	maps.Copy(dst, src)
	return dst
}

func (m *mockAlert) withMsg(msg string) *mockAlert {
	m.MetaData["message"] = msg
	return m
}

type mockHandler struct {
	receivedAlerts []Alert
}

func (m *mockHandler) Handle(_ context.Context, alerts ...Alert) {
	m.receivedAlerts = append(m.receivedAlerts, alerts...)
}

func newMockAlert(id string, severity SeverityLevel, start time.Time, end time.Time, sendAt ...time.Time) *mockAlert {
	sendAtTime := start
	if len(sendAt) > 0 {
		sendAtTime = sendAt[0]
	}

	return &mockAlert{
		Alert: &Alert{
			SourceID: id,
			From:     start,
			To:       end,
			Severity: severity,
			MetaData: make(map[string]string),
		},
		sendAt: sendAtTime,
	}
}

type resultAlert struct {
	id       string
	severity SeverityLevel
	up       bool
}

func newResultAlert(id string, severity SeverityLevel, up bool) *resultAlert {
	return &resultAlert{
		id:       id,
		severity: severity,
		up:       up,
	}
}

func createSender(window time.Duration, alerts []*mockAlert) <-chan []Alert {
	ch := make(chan []Alert)

	trueAlerts := make([]*mockAlert, 0, len(alerts)*2)

	for _, alert := range alerts {
		initialAlert := &mockAlert{
			Alert: &Alert{
				SourceID: alert.SourceID,
				From:     alert.From,
				Severity: alert.Severity,
				MetaData: cloneMeta(alert.MetaData),
			},
			sendAt: alert.sendAt,
		}

		downAlert := &mockAlert{
			Alert: &Alert{
				SourceID: alert.SourceID,
				From:     alert.From,
				To:       alert.To,
				Severity: alert.Severity,
				MetaData: cloneMeta(alert.MetaData),
			},
			sendAt: alert.To,
		}

		trueAlerts = append(trueAlerts, initialAlert.withMsg("UP"))
		trueAlerts = append(trueAlerts, downAlert.withMsg("DOWN"))
	}

	sort.Slice(trueAlerts, func(i, j int) bool {
		return trueAlerts[i].sendAt.Before(trueAlerts[j].sendAt)
	})

	go func() {
		defer close(ch)

		if len(trueAlerts) == 0 {
			return
		}

		winStart := trueAlerts[0].sendAt.Truncate(window)
		winEnd := winStart.Add(window)

		var batch []Alert

		flush := func() {
			if len(batch) > 0 {
				fmt.Println("--------flushing batch--------")
				for _, alert := range batch {
					fmt.Println(alert)
				}

				ch <- batch
				batch = nil
			}
		}

		for _, alert := range trueAlerts {
			for !alert.sendAt.Before(winEnd) {
				winStart = winEnd
				winEnd = winStart.Add(window)
				flush()
				fmt.Println("\n** New window start: ", winStart.Unix(), " **")
				time.Sleep(time.Until(winEnd))
			}

			if alert.From.Before(winStart) {
				alert.From = winStart
			}

			batch = append(batch, *alert.Alert)
		}
		flush()
	}()

	return ch
}

func TestGroupRecurringMiddleware(t *testing.T) {
	base := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	testcases := []struct {
		name           string
		inputAlerts    []*mockAlert
		expectedAlerts []*resultAlert
		ignorePeriod   time.Duration
	}{
		{
			name: "single alert",
			inputAlerts: []*mockAlert{
				newMockAlert("1", SeverityLow, base, base.Add(7*time.Minute)),
			},
			expectedAlerts: []*resultAlert{
				newResultAlert("1", SeverityLow, UP),
				newResultAlert("1", SeverityLow, DOWN),
			},
			ignorePeriod: time.Minute,
		},
		{
			name: "no group",
			inputAlerts: []*mockAlert{
				newMockAlert("1", SeverityLow, base, base.Add(7*time.Minute)),
				newMockAlert("1", SeverityLow, base.Add(16*time.Minute), base.Add(22*time.Minute)),
			},
			expectedAlerts: []*resultAlert{
				newResultAlert("1", SeverityLow, UP),
				newResultAlert("1", SeverityLow, DOWN),
				newResultAlert("1", SeverityLow, UP),
				newResultAlert("1", SeverityLow, DOWN),
			},
			ignorePeriod: 2 * time.Minute,
		},
		{
			name: "recurring alerts",
			inputAlerts: []*mockAlert{
				newMockAlert("1", SeverityLow, base, base.Add(3*time.Minute)),
				newMockAlert("1", SeverityLow, base.Add(4*time.Minute), base.Add(5*time.Minute)),
				newMockAlert("1", SeverityLow, base.Add(6*time.Minute), base.Add(7*time.Minute)),
			},
			expectedAlerts: []*resultAlert{
				newResultAlert("1", SeverityLow, UP),
				newResultAlert("1", SeverityLow, DOWN),
			},
			ignorePeriod: 5 * time.Minute,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				store := NewMockLocalCacheRepository()
				middleware := Middlewares(
					GroupRecurringMiddleware(cache.Typed[time.Time](store), tc.ignorePeriod),
				)
				mockHandler := &mockHandler{}
				handler := middleware(mockHandler)

				senderCh := createSender(5*time.Minute, tc.inputAlerts)
				for batch := range senderCh {
					handler.Handle(context.Background(), batch...)
				}

				assert.Equal(t, len(tc.expectedAlerts), len(mockHandler.receivedAlerts))

				for i, expected := range tc.expectedAlerts {
					assert.Equal(t, expected.id, mockHandler.receivedAlerts[i].SourceID)
					assert.Equal(t, expected.severity, mockHandler.receivedAlerts[i].Severity)

					up := mockHandler.receivedAlerts[i].To.IsZero()
					assert.Equal(t, up, expected.up)
				}
			})
		})
	}
}

func TestDeduplicationMiddleware(t *testing.T) {
	base := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	testcases := []struct {
		name        string
		inputAlerts []*mockAlert
		expected    []*resultAlert
	}{
		{
			name: "duplicate without finisher",
			inputAlerts: []*mockAlert{
				newMockAlert("1", SeverityLow, base, base.Add(13*time.Minute)),
				newMockAlert("1", SeverityLow, base, base.Add(13*time.Minute), base.Add(7*time.Minute)),
			},
			expected: []*resultAlert{
				newResultAlert("1", SeverityLow, UP),
				newResultAlert("1", SeverityLow, DOWN),
			},
		},
		{
			name: "duplicate after finisher",
			inputAlerts: []*mockAlert{
				newMockAlert("1", SeverityLow, base, base.Add(11*time.Minute)),
				newMockAlert("1", SeverityLow, base, base.Add(11*time.Minute), base.Add(13*time.Minute)),
			},
			expected: []*resultAlert{
				newResultAlert("1", SeverityLow, UP),
				newResultAlert("1", SeverityLow, DOWN),
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				store := NewMockLocalCacheRepository()
				middleware := Middlewares(
					DeduplicationMiddleware(cache.Typed[Alert](store), time.Hour),
				)
				mockHandler := &mockHandler{}
				handler := middleware(mockHandler)

				senderCh := createSender(5*time.Minute, testcase.inputAlerts)
				for batch := range senderCh {
					handler.Handle(context.Background(), batch...)
				}

				assert.Equal(t, len(testcase.expected), len(mockHandler.receivedAlerts))

				for i, expected := range testcase.expected {
					assert.Equal(t, expected.id, mockHandler.receivedAlerts[i].SourceID)
					assert.Equal(t, expected.severity, mockHandler.receivedAlerts[i].Severity)

					up := false
					if mockHandler.receivedAlerts[i].To.IsZero() {
						up = true
					}

					assert.Equal(t, expected.up, up)
				}
			})
		})
	}
}
