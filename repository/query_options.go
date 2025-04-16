package repository

import "go.mongodb.org/mongo-driver/bson"

type QueryOptions struct {
	Archived bool
	Fields   []string // option for read operations, empty slice means all
	Sort     SortOpt
	Filter   interface{}
	Limit    int64
	Offset   int64
}

type SortField struct {
	Field string
	Desc  bool
}

type SortOpt []SortField

func Sort() SortOpt { return nil }

func (o SortOpt) Asc(field string) SortOpt {
	return append(o, SortField{Field: field})
}

func (o SortOpt) Desc(field string) SortOpt {
	return append(o, SortField{Field: field, Desc: true})
}

type ModifyOptions struct {
	DuplicateFilter interface{}
	QueryOptions
}

func BuildDuplicateFilter[T any](fieldName string, f func(t T) interface{}, vals ...T) *ModifyOptions {
	if len(vals) == 0 {
		return nil
	}

	filter := make([]any, 0, len(vals))
	for _, val := range vals {
		filter = append(filter, f(val))
	}
	return &ModifyOptions{DuplicateFilter: bson.M{fieldName: bson.M{"$in": filter}}}
}
