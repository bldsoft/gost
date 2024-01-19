package clickhouse

import (
	"context"
	"fmt"

	"github.com/bldsoft/gost/utils/exporter"
)

type ExporterConfig struct {
	TableName string
	exporter.BufferedExporterConfig
}

type Exporter[T any] struct {
	storage     *Storage
	bufExporter *exporter.BufferedExporter[T]
	batchInsert *Batch
}

func NewExporter[T any](storage *Storage, cfg ExporterConfig) *Exporter[T] {
	e := &Exporter[T]{
		storage: storage,
	}
	e.initBatch(cfg.TableName)
	e.bufExporter = exporter.NewBuffered[T](
		exporter.Func(e.export),
		cfg.BufferedExporterConfig,
	)
	return e
}

func (e *Exporter[T]) initBatch(table string) *Exporter[T] {
	insert := fmt.Sprintf("INSERT INTO %s", table)

	batch, err := e.storage.PrepareStaticBatch(insert)
	if err != nil {
		panic(err)
	}

	e.batchInsert = batch
	return e
}

func (e *Exporter[T]) export(items ...T) (int, error) {
	if !e.storage.IsReady() {
		return 0, ErrLogDbNotReady
	}

	for _, item := range items {
		if err := e.batchInsert.Append(item); err != nil {
			return 0, err
		}
	}

	if err := e.batchInsert.Send(); err != nil {
		return 0, err
	}

	return len(items), nil
}

func (e *Exporter[T]) Export(items ...T) (n int, err error) {
	return e.bufExporter.Export(items...)
}

func (e *Exporter[T]) Run() error {
	return e.bufExporter.Run()
}

func (e *Exporter[T]) Stop(ctx context.Context) error {
	return e.bufExporter.Stop(ctx)
}
