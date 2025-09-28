package alert

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/cache/ristretto"
	"github.com/stretchr/testify/assert"
)

type mockAlert struct {
	Alert
}

func (m *mockAlert) withMsg(msg string) *mockAlert {
	m.MetaData["message"] = msg
	return m
}

type mockHandler struct {
	receivedAlerts []Alert
}

func (m *mockHandler) Handle(ctx context.Context, alerts ...Alert) {
	m.receivedAlerts = append(m.receivedAlerts, alerts...)
}

func newAlert(id string, start time.Time, end ...time.Time) *mockAlert {
	var endTime time.Time
	if len(end) > 0 {
		endTime = end[0]
	}

	return &mockAlert{
		Alert: Alert{
			SourceID: id,
			From:     start,
			To:       endTime,
			MetaData: make(map[string]string),
		},
	}
}

func createSender(window time.Duration, alerts []mockAlert) <-chan mockAlert {
	ch := make(chan mockAlert)
	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].From.Before(alerts[j].From)
	})

	go func() {
		defer close(ch)

		splitStart := time.Now()
		splitEnd := splitStart.Add(window)

		for splitStart.Before(alerts[len(alerts)-1].To) {
			for _, alert := range alerts {
				if alert.From.After(splitStart) && alert.From.Before(splitEnd) {
					//send alert
				}
			}
		}
	}()

	return ch
}

func TestDeduplicationMiddleware(t *testing.T) {
	testcases := []struct {
		name        string
		inputAlerts []*mockAlert
		outputCount int
	}{
		{
			name: "no duplicates",
			inputAlerts: []*mockAlert{
				newAlert("1", time.Now(), time.Now().Add(5*time.Minute)),
				newAlert("2", time.Now().Add(time.Minute), time.Now().Add(5*time.Minute)),
				newAlert("1", time.Now().Add(10*time.Minute), time.Now().Add(15*time.Minute)),
			},
			outputCount: 3, //TODO: check that they are correct
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			store := cache.Typed[Alert](ristretto.NewRepository("{}"))
			middleware := Middlewares(
				DeduplicationMiddleware(store, time.Second),
			)

			handler := &mockHandler{}
			wrappedHandler := middleware(handler)

			wrappedHandler.Handle(context.Background(), testcase.inputAlerts...)

			for i, alert := range handler.receivedAlerts {
				assert.Equal(t, testcase.outputMessages[i], alert.MetaData["message"])
			}
		})
	}
}
