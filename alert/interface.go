package alert

import (
	"context"
	"slices"
	"time"

	"github.com/bldsoft/gost/alert/notify"
	"github.com/bldsoft/gost/utils/poly"
)

//go:generate go run github.com/dmarkham/enumer -gqlgen --json -transform=lower -type SeverityLevel --output severity_level.go --trimprefix "SeverityLevel"
type SeverityLevel int

const (
	SeverityLevelLow SeverityLevel = iota
	SeverityLevelMedium
	SeverityLevelHigh
	SeverityLevelCritical
)

type Alert struct {
	SourceID string `bson:"sourceID" json:"sourceID"`

	Severity  SeverityLevel                `bson:"severity" json:"severity"`
	From      time.Time                    `bson:"from" json:"from"`
	To        time.Time                    `bson:"to,omitempty" json:"to,omitempty"`
	Receivers []poly.Poly[notify.Receiver] `bson:"receivers" json:"receivers"`

	MetaData map[string]any `bson:"metadata" json:"metadata"`
}

func (a Alert) AddMetaData(key string, value any) Alert {
	if a.MetaData == nil {
		a.MetaData = make(map[string]any)
	}
	a.MetaData[key] = value
	return a
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

type Repository interface {
	CreateAlert(ctx context.Context, alerts Alert) error
	GetAlert(ctx context.Context, id string, level SeverityLevel) (Alert, error)
	UpdateAlert(ctx context.Context, id string, level SeverityLevel, newAlert Alert) error
}
