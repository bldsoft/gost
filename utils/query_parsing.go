package utils

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/gorilla/schema"
)

var (
	decoder = schema.NewDecoder()
	encoder = schema.NewEncoder()
)

func init() {
	decoder.IgnoreUnknownKeys(true)
	decoder.RegisterConverter(time.Time{}, func(s string) reflect.Value {
		ts, _ := strconv.ParseInt(s, 10, 64)
		return reflect.ValueOf(time.Unix(ts, 0))
	})
	encoder.RegisterEncoder(time.Time{}, func(v reflect.Value) string {
		return strconv.FormatInt(v.Interface().(time.Time).Unix(), 10)
	})

	var d time.Duration
	decoder.RegisterConverter(d, func(s string) reflect.Value {
		dur, _ := time.ParseDuration(s)
		return reflect.ValueOf(dur)
	})
	encoder.RegisterEncoder(d, func(v reflect.Value) string {
		return time.Duration(v.Int()).String()
	})
}

func FromRequest[T any](r *http.Request) (*T, error) {
	return FromQuery[T](r.URL.Query())
}

func FromQuery[T any](query url.Values) (*T, error) {
	var obj T
	return &obj, decoder.Decode(&obj, query)
}

func Query[T any](obj T) url.Values {
	query := make(url.Values)
	if err := encoder.Encode(obj, query); err != nil {
		panic(err)
	}
	return query
}
