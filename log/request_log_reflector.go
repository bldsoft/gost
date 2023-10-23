package log

import (
	"fmt"
	"reflect"
	"strings"
)

const tag_name = "requestinfo"

func PrepareQuery(ri any, into string) (string, error) {
	fields := []string{}
	f := func(field reflect.StructField, _ reflect.Value) {
		tag, ok := tagCheck(field)
		if !ok {
			return
		}
		fields = append(fields, tag)
	}
	if err := traverse(ri, f); err != nil {
		return "", err
	}

	values := strings.Repeat("?,", len(fields))
	return `INSERT INTO ` + into + `(` + strings.Join(fields, ",") + `)
			VALUES (` + values[:len(values)-1] + `)`,
		nil
}

func Values(ri any) ([]any, error) {
	values := []any{}
	f := func(field reflect.StructField, value reflect.Value) {
		if _, ok := tagCheck(field); !ok {
			return
		}
		values = append(values, value.Interface())
	}

	if err := traverse(ri, f); err != nil {
		return nil, err
	}

	return values, nil
}

func traverse(ri any, f func(field reflect.StructField, value reflect.Value)) error {
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
			_ = traverse(v.Field(i).Interface(), f)
			continue
		}
		f(field, v.Field(i))
	}

	return nil
}

func tagCheck(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get(tag_name)
	return tag, len(tag) > 0
}
