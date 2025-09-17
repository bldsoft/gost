package test

import (
	"context"
	"time"

	"github.com/bldsoft/gost/alert"
)

type ThresholdChecker struct{}

func Check(ctx context.Context, thresholds []alert.Threshold) (fires [][]alert.ThresholdFire, err error) {
	res := make([][]alert.ThresholdFire, len(thresholds))
	for i := range len(res) {
		res[i] = append(res[i], alert.ThresholdFire{
			From: time.Now().Add(-10 * time.Second),
			To:   time.Now().Add(-5 * time.Second),
		})
	}
	return res, nil
}
