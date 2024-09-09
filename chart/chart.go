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

func BuildChartValues(start, end time.Time, step int64, times []time.Time, values []float64) []*SeriesValue {
	start = time.Unix(start.Unix()-start.Unix()%step, 0)
	res := make([]*SeriesValue, 0, (end.Unix()-start.Unix())/step)
	for t := start.Unix(); t < end.Unix(); t += step {
		v := &SeriesValue{Time: t}
		res = append(res, v)
		if len(times) > 0 && times[0].Unix() == t {
			v.Value = &values[0]
			times, values = times[1:], values[1:]
		}
	}
	return res
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
