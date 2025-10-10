package alert

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"testing"
	"testing/synctest"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/cache/bigcache"
	"github.com/stretchr/testify/assert"
)

const (
	UP       = "UP"
	DOWN     = "DOWN"
	HAPPENED = "HAPPENED" //up and down was in the same time window
)

type mockRepository struct {
	Store map[string]Alert
}

func (r *mockRepository) CreateAlerts(ctx context.Context, alerts ...Alert) error {
	for _, alert := range alerts {
		key := fmt.Sprintf("%s%d", alert.SourceID, alert.Severity)
		r.Store[key] = alert
	}
	return nil
}

func (r *mockRepository) GetAlert(ctx context.Context, id string, level SeverityLevel) (Alert, error) {
	key := fmt.Sprintf("%s%d", id, level)
	alert, ok := r.Store[key]
	if !ok {
		return Alert{}, errors.New("not found")
	}
	return alert, nil
}

func (r *mockRepository) UpdateAlert(ctx context.Context, id string, level SeverityLevel, newAlert Alert) error {
	key := fmt.Sprintf("%s%d", id, level)
	_, ok := r.Store[key]
	if !ok {
		return errors.New("not found")
	}

	r.Store[key] = newAlert
	return nil
}

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
	base := fmt.Sprintf("%s, %s: %s", m.SourceID, m.SeverityString(), m.From)
	if !m.To.IsZero() {
		base = fmt.Sprintf("%s-%s", base, m.To)
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
	status   string
}

func newResultAlert(id string, severity SeverityLevel, status string) *resultAlert {
	return &resultAlert{
		id:       id,
		severity: severity,
		status:   status,
	}
}

func createSender(window time.Duration, alerts []*mockAlert) <-chan []Alert {
	ch := make(chan []Alert)

	trueAlerts := make([]*mockAlert, 0, len(alerts)*2)

	for _, alert := range alerts {
		startWindow := alert.From.Truncate(window)
		endWindow := alert.To.Truncate(window)

		if startWindow.Equal(endWindow) {
			happenedAlert := &mockAlert{
				Alert: &Alert{
					SourceID: alert.SourceID,
					From:     alert.From,
					To:       alert.To,
					Severity: alert.Severity,
					MetaData: cloneMeta(alert.MetaData),
				},
				sendAt: alert.From,
			}
			trueAlerts = append(trueAlerts, happenedAlert.withMsg("HAPPENED"))
		} else {
			upAlert := &mockAlert{
				Alert: &Alert{
					SourceID: alert.SourceID,
					From:     alert.From,
					Severity: alert.Severity,
					MetaData: cloneMeta(alert.MetaData),
				},
				sendAt: alert.From,
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

			trueAlerts = append(trueAlerts, upAlert.withMsg("UP"))
			trueAlerts = append(trueAlerts, downAlert.withMsg("DOWN"))
		}
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
				time.Sleep(time.Until(winEnd))
				fmt.Println("\n** New window start: ", time.Now(), " **")
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
				newMockAlert("1", SeverityLow, base.Add(4*time.Minute), base.Add(6*time.Minute)),
				newMockAlert("1", SeverityLow, base.Add(6*time.Minute), base.Add(9*time.Minute)),
			},
			expectedAlerts: []*resultAlert{
				newResultAlert("1", SeverityLow, HAPPENED),
				newResultAlert("1", SeverityLow, HAPPENED),
			},
			ignorePeriod: 5 * time.Minute,
		},
		{
			name: "recurring alerts with delayed end",
			inputAlerts: []*mockAlert{
				newMockAlert("1", SeverityLow, base, base.Add(3*time.Minute)),
				newMockAlert("1", SeverityLow, base.Add(4*time.Minute), base.Add(6*time.Minute)),
				newMockAlert("1", SeverityLow, base.Add(6*time.Minute), base.Add(12*time.Minute)),
			},
			expectedAlerts: []*resultAlert{
				newResultAlert("1", SeverityLow, HAPPENED),
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
					GroupRecurringMiddleware(cache.Typed[Alert](store), tc.ignorePeriod),
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
					assert.Equal(t, expected.status, mockHandler.receivedAlerts[i].MetaData["message"])
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
		{
			name: "fast duplicates",
			inputAlerts: []*mockAlert{
				newMockAlert("1", SeverityLow, base, base.Add(4*time.Minute)),
				newMockAlert("1", SeverityMedium, base.Add(time.Minute), base.Add(3*time.Minute)),
				newMockAlert("1", SeverityLow, base.Add(7*time.Minute), base.Add(13*time.Minute)),
			},
			expected: []*resultAlert{
				newResultAlert("1", SeverityLow, HAPPENED),
				newResultAlert("1", SeverityMedium, HAPPENED),
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
					assert.Equal(t, expected.status, mockHandler.receivedAlerts[i].MetaData["message"])
				}
			})
		})
	}
}

func TestMiddlewares(t *testing.T) {
	base := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	testcases := []struct {
		name         string
		ignorePeriod time.Duration
		inputAlerts  []*mockAlert
		expected     []*mockAlert
	}{
		{
			name:         "plain alerts of different severity",
			ignorePeriod: time.Second,
			inputAlerts: []*mockAlert{
				newMockAlert("1", SeverityLow, base, base.Add(9*time.Minute)),
				newMockAlert("1", SeverityMedium, base.Add(3*time.Minute), base.Add(8*time.Minute)),
				newMockAlert("1", SeverityHigh, base.Add(4*time.Minute), base.Add(7*time.Minute)),
				newMockAlert("1", SeverityLow, base.Add(11*time.Minute), base.Add(13*time.Minute)),
			},
			expected: []*mockAlert{
				newMockAlert("1", SeverityLow, base, base.Add(9*time.Minute)),
				newMockAlert("1", SeverityMedium, base.Add(3*time.Minute), base.Add(8*time.Minute)),
				newMockAlert("1", SeverityHigh, base.Add(4*time.Minute), base.Add(7*time.Minute)),
				newMockAlert("1", SeverityLow, base.Add(11*time.Minute), base.Add(13*time.Minute)),
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				store := NewMockLocalCacheRepository()
				rep := new(mockRepository)

				middleware := Middlewares(
					DeduplicationMiddleware(cache.Typed[Alert](store), time.Hour),
					GroupRecurringMiddleware(cache.Typed[Alert](store), testcase.ignorePeriod),
					ArchivingMiddleware(rep),
				)
				mockHandler := &mockHandler{}
				handler := middleware(mockHandler)

				senderCh := createSender(5*time.Minute, testcase.inputAlerts)
				for batch := range senderCh {
					handler.Handle(context.Background(), batch...)
				}

				assert.Equal(t, len(testcase.expected), len(rep.Store))

				stored := slices.Collect(maps.Values(rep.Store))
				slices.SortFunc(stored, func(a, b Alert) int {
					if a.From.Before(b.From) {
						return -1
					}
					return 1
				})

				for i, expected := range testcase.expected {
					assert.Equal(t, expected.SourceID, stored[i].SourceID)
					assert.Equal(t, expected.Severity, stored[i].Severity)
					assert.Equal(t, expected.From, stored[i].From)
					assert.Equal(t, expected.To, stored[i].To)
				}
			})
		})
	}
}
