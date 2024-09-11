package clickhouse

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bldsoft/gost/entity/chart"
)

func (r *BaseRepository) buildChartValues(start, end time.Time, step time.Duration, times []time.Time, values []float64) []*chart.SeriesValue {
	// start = time.Unix(start.Unix()-start.Unix()%step, 0)
	// res := make([]*chart.SeriesValue, 0, (end.Unix()-start.Unix())/step)
	// for t := start.Unix(); t < end.Unix(); t += step {
	// 	v := &chart.SeriesValue{Time: t}
	// 	res = append(res, v)
	// 	if len(times) > 0 && times[0].Unix() == t {
	// 		v.Value = &values[0]
	// 		times, values = times[1:], values[1:]
	// 	}
	// }
	// return res

	start = start.Add(-time.Duration(start.UnixNano()) % step)
	res := make([]*chart.SeriesValue, 0, int(end.Sub(start)/step))
	for t := start; t.Before(end); t = t.Add(step) {
		v := &chart.SeriesValue{Time: t.Unix()}
		res = append(res, v)
		if len(times) > 0 && times[0].Equal(t) {
			v.Value = &values[0]
			times, values = times[1:], values[1:]
		}
	}
	return res
}

func (r *BaseRepository) GetChartValues(ctx context.Context, subQuery sq.SelectBuilder, from, to time.Time, step time.Duration) (*chart.SeriesData, error) {
	query := sq.Select().
		Column(labelColumn).
		Column("groupArray("+timeColumn+") "+timesColumn).
		Column("groupArray(toFloat64("+valueColumn+")) "+valuesColumn).
		Column("min(toFloat64("+valueColumn+")) min").
		Column("max(toFloat64("+valueColumn+")) max").
		Column("avg(toFloat64("+valueColumn+")) avg").
		Column("sum(toFloat64("+valueColumn+")) sum").
		FromSelect(subQuery, "interval_data").
		GroupBy(labelColumn)

	data, err := chart.NewSeriesData(from, to, step)
	if err != nil {
		return nil, err
	}

	rows, err := r.RunSelect(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			lv     chart.SeriesValues
			times  []time.Time
			values []float64
		)
		if err := rows.Scan(&lv.Label, &times, &values, &lv.Min, &lv.Max, &lv.Avg, &lv.Sum); err != nil {
			return nil, err
		}
		lv.Data = r.buildChartValues(from, to, step, times, values)
		data.Values = append(data.Values, &lv)
	}
	return data, nil
}
