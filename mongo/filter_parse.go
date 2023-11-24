package mongo

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

func RecursiveParse(filter bson.M, t interface{}, prefix string) {
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
			RecursiveParse(filter, v.Interface(), prefix+field.Tag.Get("bson")+".")
		case reflect.Pointer:
			tag = prefix + tag
			filter[tag] = v.Elem().Interface()
		case reflect.Array, reflect.Slice:
			filter[tag] = bson.M{"$in": v.Interface()}
		default:
			continue
		}
	}
}
