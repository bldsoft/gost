package alert

import (
	"context"
	"slices"
	"time"

	"github.com/bldsoft/gost/alert/notify"
	"github.com/bldsoft/gost/utils/poly"
)

// ENUM(SeverityLow,SeverityMedium,SeverityHigh,SeverityCritical)
type SeverityLevel int

type Alert struct {
	SourceID string

	Severity  SeverityLevel
	From, To  time.Time
	Receivers []poly.Poly[notify.Receiver]

	MetaData map[string]string
}

type Handler interface {
	Handle(ctx context.Context, alerts ...Alert)
}

type Middleware func(Handler) Handler

type HandlerFunc func(ctx context.Context, alerts ...Alert)

func (f HandlerFunc) Handle(ctx context.Context, alerts ...Alert) {
	f(ctx, alerts...)
}

func Middlewares(middlewares ...Middleware) Middleware {
	return func(next Handler) Handler {
		for _, m := range slices.Backward(middlewares) {
			next = m(next)
		}
		return next
	}
}

type Notifier interface {
	Send(ctx context.Context, alerts Alert) error
}

type Source interface {
	EvaluateAlerts(ctx context.Context) (alerts []Alert, next time.Time, err error)
}

type SourceFunc func(ctx context.Context) ([]Alert, time.Time, error)

func (f SourceFunc) EvaluateAlerts(ctx context.Context) ([]Alert, time.Time, error) {
	return f(ctx)
}
