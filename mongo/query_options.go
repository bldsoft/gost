package mongo

import (
	"reflect"
	"strings"
	"unsafe"

	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
)

func ParseQueryOptions[T any](q *repository.QueryOptions[T]) bson.M {
	if q == nil {
		return bson.M{}
	}

	filter := bson.M{}
	f := q.Filter
	recursiveParse(filter, f, []string{})
	return filter
}

func recursiveParse(filter bson.M, t interface{}, parents []string) {
	fv := reflect.Indirect(reflect.ValueOf(t))
	ft := fv.Type()

	for i, limit := 0, fv.NumField(); i < limit; i++ {
		field := ft.Field(i)

		v := fv.FieldByName(field.Name)
		if v.IsZero() {
			continue
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			var _parents []string
			copy(parents, _parents)
			recursiveParse(filter, v.Elem(), append(_parents, field.Tag.Get("bson")))
		case reflect.Pointer:
			tag := field.Tag.Get("bson")
			tag = strings.Join(parents, ".") + tag
			filter[tag] = unsafe.Pointer(v.Pointer())
			el := v.Elem()
			filter[tag] = el
		default:
			continue
		}

	}
}
