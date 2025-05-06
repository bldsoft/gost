package clickhouse

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/bldsoft/gost/utils/exporter"
)

type exporterBatch[T any] struct {
	batch *Batch
	n     int
}

func newExporterBatch[T any](storage *Storage, table string) *exporterBatch[T] {
	columns := strings.Join(columnNames[T](), ",")
	insert := fmt.Sprintf("INSERT INTO %s (%s) VALUES", table, columns)
	batch, err := storage.PrepareStaticBatch(insert)
	if err != nil {
		panic(err)
	}
	return &exporterBatch[T]{batch: batch}
}

func (e *exporterBatch[T]) Send() error {
	if err := e.batch.Send(); err != nil {
		return err
	}
	e.n = 0
	return nil
}

func (e *exporterBatch[T]) Len() int {
	return e.n
}

func (e *exporterBatch[T]) Add(items ...T) (n int, err error) {
	for i, item := range items {
		if err := e.batch.Append(item); err != nil {
			return i, err
		}
		e.n++
	}
	return len(items), nil
}

func (e *exporterBatch[T]) Reset() error {
	if err := e.batch.reset(); err != nil {
		return err
	}
	e.n = 0
	return nil
}

var _ exporter.Data[int] = (*exporterBatch[int])(nil)

func columnNames[T any]() []string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return columnNamesFromType(t)
}

// https://github.com/ClickHouse/clickhouse-go/blob/main/struct_map.go
func columnNamesFromType(t reflect.Type) []string {
	var keys []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := f.Name

		if tn := f.Tag.Get("ch"); len(tn) != 0 {
			name = tn
		}
		if name == "-" || (len(f.PkgPath) != 0 && !f.Anonymous) {
			continue
		}

		if f.Anonymous && f.Type.Kind() != reflect.Ptr {
			subKeys := columnNamesFromType(f.Type)
			keys = append(keys, subKeys...)
		} else {
			keys = append(keys, name)
		}
	}
	return keys
}
