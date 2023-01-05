package repository

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

const (
	SchemaTag = "schema"
	Archived  = "archived"
)

type QueryOptions[F any] struct {
	Archived bool
	Fields   []string // option for read operations, empty slice means all
	Filter   F
}

func ParseURLQuery[F any](q url.Values) (*QueryOptions[F], error) {
	opts := QueryOptions[F]{}
	if archived_str := q.Get(Archived); archived_str != "" {
		if val, err := strconv.ParseBool(archived_str); err != nil {
			return nil, err
		} else {
			opts.Archived = val
			q.Del(Archived)
		}

	}
	for k := range q {
		if err := fillOptions(&opts, k, q.Get(k)); err != nil {
			return nil, err
		}
	}
	return &opts, nil
}

func fillOptions(f interface{}, schema string, val string) error {
	var (
		childs = []interface{}{}
		found  bool
	)

	traverseFields := func(f interface{}) error {
		fv := reflect.Indirect(reflect.ValueOf(f))
		ft := fv.Type()

		for i, limit := 0, fv.NumField(); i < limit && !found; i++ {
			field := ft.Field(i)
			v := fv.FieldByName(field.Name)
			kind := field.Type.Kind()
			if kind == reflect.Struct {
				childs = append(childs, v.Interface())
			}

			if field.Tag.Get(SchemaTag) != schema {
				continue
			}

			return set(&v, val)

		}
		return nil
	}

	if err := traverseFields(f); err != nil {
		return err
	}

	for len(childs) != 0 {
		if err := traverseFields(childs[0]); err != nil {
			return err
		}
		childs = childs[1:]
	}

	return nil
}

func set(v *reflect.Value, val string) error {
	if kind := v.Kind(); kind != reflect.Pointer {
		return fmt.Errorf("%s is not a pointer", kind)
	}
	switch v.Type().Elem().Kind() {
	case reflect.String:
		v.Set(reflect.ValueOf(&val))
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(&b))
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(&f))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(&i))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		v.Set(reflect.ValueOf(&i))
	default:
		return fmt.Errorf("unsopported conversion")
	}
	return nil
}
