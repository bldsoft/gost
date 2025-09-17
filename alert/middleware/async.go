package middleware

import (
	"context"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/utils/errgroup"
)

func Async(next alert.AlertHandler) alert.AlertHandler {
	return alert.AlertHandlerFunc(func(ctx context.Context, alerts ...alert.Alert) {
		var errGroup errgroup.Group
		errGroup.Go(func() error {
			next.Handle(ctx, alerts...)
			return nil
		})
	})
}
