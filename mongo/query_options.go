package mongo

import (
	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
)

func ParseQueryOptions(q *repository.QueryOptions) bson.M {
	if q == nil {
		return bson.M{}
	}

	filter := bson.M{}
	f := q.Filter
	RecursiveParse(filter, f, "")
	return filter
}
