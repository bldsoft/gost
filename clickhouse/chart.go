package clickhouse

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bldsoft/gost/entity/stat"
)

func (r *BaseRepository) buildChartValues(start, end time.Time, step time.Duration, times []time.Time, values []float64) []*stat.SeriesValue {
	start = start.Add(-time.Duration(start.UnixNano()) % step)
	res := make([]*stat.SeriesValue, 0, int(end.Sub(start)/step))
	for t := start; t.Before(end); t = t.Add(step) {
		v := &stat.SeriesValue{Time: t.Unix()}
		res = append(res, v)
		if len(times) > 0 && times[0].Equal(t) {
			v.Value = &values[0]
			times, values = times[1:], values[1:]
		}
	}
	return res
}

func (r *BaseRepository) GetChartValues(ctx context.Context, subQuery sq.SelectBuilder, from, to time.Time, step time.Duration) (*stat.SeriesData, error) {
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

	data, err := stat.NewSeriesData(from, to, step)
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
			lv     stat.SeriesValues
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
