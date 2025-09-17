package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"testing/synctest"

	"github.com/bldsoft/gost/alert"
	"github.com/bldsoft/gost/alert/middleware"
	"github.com/bldsoft/gost/alert/notify"
	"github.com/bldsoft/gost/utils/errgroup"
	"github.com/stretchr/testify/require"
)

func TestAlert(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*11)
		defer cancel()

		alertGen := NewAlertGen(10, time.Second)
		group, groupWait := middleware.Group(ctx, middleware.GroupConfig{
			GroupInterval: 2 * time.Second,
		})
		_ = group
		var notifier notify.SliceCollector

		alertManager := alert.NewAlertManager()

		alertManager.AddAlertProcessor(alert.Processor{
			Cfg: alert.ProcessorCfg{
				CheckInterval: time.Second,
			},
			Checker: alertGen,
			AlertHandler: alert.Middlewares(
				middleware.Deduplication(10*time.Second),
				group,
			)(notify.NewHandler(notify.Sync(&notifier))),
		})

		var errGroup errgroup.Group
		errGroup.Go(func() error {
			alertGen.Start(ctx)
			return nil
		})
		errGroup.Go(func() error {
			return alertManager.Run(ctx)
		})

		_ = errGroup.Wait()
		groupWait()
		require.Equal(t, 1, len(notifier.Alerts()), func() string {
			res := make([]string, 0, len(notifier.Alerts()))
			for _, a := range notifier.Alerts() {
				res = append(res, a.ID)
			}
			return strings.Join(res, ",")
		}())
	})
}
