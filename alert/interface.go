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
	ID string

	Severity  SeverityLevel
	From, To  time.Time
	Notifiers []Notifier

	MetaData map[string]string
}

type AlertHandler interface {
	Handle(ctx context.Context, alerts ...Alert)
}

type Middleware func(AlertHandler) AlertHandler

type AlertHandlerFunc func(ctx context.Context, alerts ...Alert)

func (f AlertHandlerFunc) Handle(ctx context.Context, alerts ...Alert) {
	f(ctx, alerts...)
}

func Middlewares(middlewares ...Middleware) Middleware {
	return func(next AlertHandler) AlertHandler {
		for _, m := range slices.Backward(middlewares) {
			next = m(next)
		}
		return next
	}
}

type Notifier interface {
	Send(ctx context.Context, alerts Alert) error
}

type Threshold struct {
	Severity  SeverityLevel
	Condition int
	Value     float64
	Period    time.Duration
	Notifiers []Notifier
}

type Checker interface {
	CheckAlerts(ctx context.Context) ([]Alert, error)
}

type CheckerFunc func(ctx context.Context) ([]Alert, error)

func (f CheckerFunc) CheckAlerts(ctx context.Context) ([]Alert, error) {
	return f(ctx)
}
