package mongo

import (
	"reflect"

	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
)

func ParseQueryOptions(q *repository.QueryOptions) bson.M {
	if q == nil {
		return bson.M{}
	}

	filter := bson.M{}
	f := q.Filter
	recursiveParse(filter, f, "")
	return filter
}

func recursiveParse(filter bson.M, t interface{}, prefix string) {
	if t == nil {
		return
	}

	fv := reflect.Indirect(reflect.ValueOf(t))
	ft := fv.Type()

	for i, limit := 0, fv.NumField(); i < limit; i++ {
		field := ft.Field(i)

		v := fv.FieldByName(field.Name)
		if v.IsZero() {
			continue
		}

		tag := field.Tag.Get("bson")

		switch field.Type.Kind() {
		case reflect.Struct:
			recursiveParse(filter, v.Interface(), prefix+field.Tag.Get("bson")+".")
		case reflect.Pointer:
			tag = prefix + tag
			filter[tag] = v.Elem().Interface()
		default:
			continue
		}

	}
}
