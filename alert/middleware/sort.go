package middleware

import (
	"cmp"
	"context"
	"slices"

	"github.com/bldsoft/gost/alert"
)

func Sort(handler alert.Handler) alert.Handler {
	return alert.HandlerFunc(func(ctx context.Context, alerts ...alert.Alert) {
		slices.SortFunc(alerts, func(a, b alert.Alert) int {
			return cmp.Or(
				a.From.Compare(b.From),
				a.To.Compare(b.To),
			)
		})
		handler.Handle(ctx, alerts...)
	})
}
