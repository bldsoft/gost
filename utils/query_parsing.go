package utils

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/schema"
)

var decoder = schema.NewDecoder()
var encoder = schema.NewEncoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
	decoder.RegisterConverter(time.Time{}, func(s string) reflect.Value {
		ts, _ := strconv.ParseInt(s, 10, 64)
		return reflect.ValueOf(time.Unix(ts, 0))
	})

	encoder.RegisterEncoder(time.Time{}, func(v reflect.Value) string {
		return strconv.FormatInt(v.Int(), 10)
	})
}

func FromRequest[T any](r *http.Request) *T {
	return FromQuery[T](r.URL.Query())
}

func FromQuery[T any](query url.Values) *T {
	var obj T
	if err := decoder.Decode(&obj, query); err != nil {
		panic(err)
	}
	return &obj
}

func Query[T any](obj T) url.Values {
	query := make(url.Values)
	if err := encoder.Encode(obj, query); err != nil {
		panic(err)
	}
	return query
}
