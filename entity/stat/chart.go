package stat

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
