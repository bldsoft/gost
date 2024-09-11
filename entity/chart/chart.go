package chart

import (
	"errors"
	"time"
)

const (
	maxChartIntervalCount = 2048
)

type SeriesData struct {
	Values []*SeriesValues `json:"values"`
}

type SeriesValues struct {
	Label string         `json:"label"`
	Max   *float64       `json:"max,omitempty"`
	Min   *float64       `json:"min,omitempty"`
	Avg   *float64       `json:"avg,omitempty"`
	Sum   *float64       `json:"sum,omitempty"`
	Data  []*SeriesValue `json:"data"`
}

type SeriesValue struct {
	Time  int64    `json:"time"`
	Value *float64 `json:"value"`
}

func NewSeriesData(from, to time.Time, step time.Duration) (*SeriesData, error) {
	var sd SeriesData
	if step <= 0 {
		return nil, errors.New("step must be a positive number")
	}

	if to.Sub(from)/step > maxChartIntervalCount {
		return nil, errors.New("too many intervals")
	}
	return &sd, nil
}
