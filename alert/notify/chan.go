package notify

import (
	"context"
	"sync"

	"github.com/bldsoft/gost/alert"
)

// for testing
type Chan chan alert.Alert

func (c Chan) Send(ctx context.Context, alert alert.Alert) error {
	c <- alert
	return nil
}

type SliceCollector []alert.Alert

func (s *SliceCollector) Send(ctx context.Context, alert alert.Alert) error {
	*s = append(*s, alert)
	return nil
}

func (s *SliceCollector) Alerts() []alert.Alert {
	return *s
}

type syncNotifier struct {
	notifier alert.Notifier
	mtx      sync.Mutex
}

func Sync(notifier alert.Notifier) alert.Notifier {
	return &syncNotifier{notifier: notifier}
}

func (s *syncNotifier) Send(ctx context.Context, alerts alert.Alert) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.notifier.Send(ctx, alerts)
}
