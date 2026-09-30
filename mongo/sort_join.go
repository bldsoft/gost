package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/bldsoft/gost/repository"
)

const sortJoinAsPrefix = "_sortJoin"

type SortJoin struct {
	From         string
	ForeignField string
	SortBy       string
	Collation    *options.Collation
}

func (r *BaseRepository[T, U]) SetSortJoin(field string, join SortJoin) {
	if r.sortJoins == nil {
		r.sortJoins = make(map[string]SortJoin)
	}
	r.sortJoins[field] = join
}

func (r *BaseRepository[T, U]) hasSortJoin(sort repository.SortOpt) bool {
	for _, sortParam := range sort {
		if _, ok := r.sortJoins[sortParam.Field]; ok {
			return true
		}
	}

	return false
}

func (r *BaseRepository[T, U]) findWithSortJoin(ctx context.Context, filter interface{}, opt *repository.QueryOptions) ([]U, error) {
	pipeline, collation := r.sortJoinPipeline(filter, opt)
	aggregateOpt := options.Aggregate()
	if collation != nil {
		aggregateOpt.SetCollation(collation)
	}

	cur, err := r.Collection().Aggregate(ctx, pipeline, aggregateOpt)
	if err != nil {
		return nil, WrapErr(err)
	}
	results := make([]U, 0)
	if err = cur.All(ctx, &results); err != nil {
		return nil, WrapErr(err)
	}

	return results, nil
}

func (r *BaseRepository[T, U]) sortJoinPipeline(filter any, opt *repository.QueryOptions) (mongo.Pipeline, *options.Collation) {
	pipeline := mongo.Pipeline{{{Key: "$match", Value: r.where(filter, opt)}}}

	var (
		sort, joined bson.D
		collation    *options.Collation
	)
	for i, sortParam := range opt.Sort {
		order := 1
		if sortParam.Desc {
			order = -1
		}

		join, ok := r.sortJoins[sortParam.Field]
		if !ok {
			sort = append(sort, bson.E{Key: sortParam.Field, Value: order})

			continue
		}

		as := fmt.Sprintf("%s%d", sortJoinAsPrefix, i)
		pipeline = append(pipeline,
			bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: join.From},
				{Key: "localField", Value: sortParam.Field},
				{Key: "foreignField", Value: join.ForeignField},
				{Key: "as", Value: as},
			}}},
			bson.D{{Key: "$addFields", Value: bson.D{
				{Key: as, Value: bson.D{{Key: "$ifNull", Value: bson.A{
					bson.D{{Key: "$arrayElemAt", Value: bson.A{"$" + as + "." + join.SortBy, 0}}},
					"",
				}}}},
			}}},
		)
		sort = append(sort, bson.E{Key: as, Value: order})
		joined = append(joined, bson.E{Key: as, Value: 0})
		if collation == nil {
			collation = join.Collation
		}
	}
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sort}})

	if opt.Offset > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: opt.Offset}})
	}
	if opt.Limit > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: opt.Limit}})
	}

	if projection := r.projection(opt); projection != nil {
		pipeline = append(pipeline, bson.D{{Key: "$project", Value: projection}})
	} else {
		pipeline = append(pipeline, bson.D{{Key: "$project", Value: joined}})
	}

	return pipeline, collation
}
