package log

import (
	"fmt"
	"reflect"
	"strings"
)

const tag_name = "requestinfo"

func PrepareQuery(ri any, into string) (string, error) {
	fields := []string{}
	f := func(field reflect.StructField, _ reflect.Value, into *[]string) {
		fields = append(fields, field.Tag.Get(tag_name))
	}
	if err := traverse[string](ri, &fields, f); err != nil {
		return "", err
	}

	values := strings.Repeat("?,", len(fields))
	return `INSERT INTO ` + into + `(` + strings.Join(fields, ",") + `)
			VALUES (` + values[:len(values)-1] + `)`,
		nil
}

func Values(ri any) ([]any, error) {
	values := []any{}
	f := func(_ reflect.StructField, value reflect.Value, into *[]any) {
		values = append(values, value.Interface())
	}

	if err := traverse[any](ri, &values, f); err != nil {
		return nil, err
	}

	return values, nil
}

func traverse[T any](ri any, into *[]T, f func(field reflect.StructField, value reflect.Value, into *[]T)) error {
	t := reflect.TypeOf(ri)
	rk := t.Kind()
	if rk == reflect.Pointer {
		t = t.Elem()
		rk = t.Kind()
	}
	if rk != reflect.Struct {
		return fmt.Errorf("not a struct: %v", rk)
	}

	v := reflect.Indirect(reflect.ValueOf(ri))
	size := t.NumField()
	for i := 0; i < size; i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct {
			_ = traverse[T](v.Field(i).Interface(), into, f)
			continue
		}
		tag := field.Tag.Get(tag_name)
		if len(tag) == 0 {
			continue
		}
		f(field, v.Field(i), into)
	}

	return nil
}
