package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/bldsoft/gost/entity/stat"
	"github.com/bldsoft/gost/log"
)

const (
	maxChartIntervalCount = 2048
)

type BaseRepository struct {
	db *Storage
}

func NewBaseRepository(storage *Storage) BaseRepository {
	return BaseRepository{db: storage}
}

func (r *BaseRepository) Storage() *Storage {
	return r.db
}

func (r *BaseRepository) RunSelect(ctx context.Context, query sq.SelectBuilder) (*sql.Rows, error) {
	r.LogQuery(ctx, query)
	return query.RunWith(r.Storage().Db).QueryContext(ctx)
}

func (r *BaseRepository) LogQuery(ctx context.Context, query sq.SelectBuilder) {
	str, params, _ := query.ToSql()
	log.FromContext(ctx).TracefWithFields(log.Fields{"params": params}, "Query: %s", str)
}

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

func (r *BaseRepository) getCustomChartValues(ctx context.Context, query sq.SelectBuilder, from, to time.Time, step time.Duration) (*stat.SeriesData, error) {
	if step <= 0 {
		return nil, errors.New("step must be a positive number")
	}

	if to.Sub(from)/step > maxChartIntervalCount {
		return nil, errors.New("too many intervals")
	}

	data := &stat.SeriesData{}
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

	return r.getCustomChartValues(ctx, query, from, to, step)
}
