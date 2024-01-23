package clickhouse

import (
	"fmt"

	"github.com/bldsoft/gost/utils/exporter"
)

type exporterBatch[T any] struct {
	batch *Batch
	n     int
}

func newExporterBatch[T any](storage *Storage, table string) *exporterBatch[T] {
	insert := fmt.Sprintf("INSERT INTO %s", table)
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
