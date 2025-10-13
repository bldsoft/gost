package alert

import (
	"context"
	"slices"
	"time"
)

type SeverityLevel int

const (
	SeverityLow SeverityLevel = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

type Alert struct {
	SourceID string

	Severity  SeverityLevel
	From, To  time.Time
	Notifiers []Notifier

	MetaData map[string]string
}

type Handler interface {
	Handle(ctx context.Context, alerts ...Alert)
}

type Middleware func(Handler) Handler

type AlertHandlerFunc func(ctx context.Context, alerts ...Alert)

func (f AlertHandlerFunc) Handle(ctx context.Context, alerts ...Alert) {
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

type Repository interface {
	CreateAlert(ctx context.Context, alerts Alert) error
	GetAlert(ctx context.Context, id string, level SeverityLevel) (Alert, error)
	UpdateAlert(ctx context.Context, id string, level SeverityLevel, newAlert Alert) error
}
