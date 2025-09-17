package alert

import (
	"context"
	"time"
)

type ThresholdChecker interface {
	Check(ctx context.Context, thresholds []Threshold) (ok [][]ThresholdFire, err error)
}

type ThresholdFire struct {
	From     time.Time
	To       time.Time
	MetaData map[string]string
}

type ThresholdAlert struct {
	ID         string
	source     ThresholdChecker
	thresholds []Threshold
}

func (t *ThresholdAlert) CheckAlerts(ctx context.Context) ([]Alert, error) {
	thresholdsFires, err := t.source.Check(ctx, t.thresholds)
	if err != nil || len(thresholdsFires) == 0 {
		return nil, err
	}

	res := make([]Alert, 0, len(thresholdsFires))
	for i, thresholdFires := range thresholdsFires {
		threshold := t.thresholds[i]
		for _, fire := range thresholdFires {
			res = append(res, Alert{
				ID:        t.ID,
				Severity:  threshold.Severity,
				From:      fire.From,
				To:        fire.To,
				Notifiers: threshold.Notifiers,
				MetaData:  fire.MetaData,
			})
		}
	}

	return res, nil
}
