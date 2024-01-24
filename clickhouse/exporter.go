package clickhouse

import (
	"context"

	"github.com/bldsoft/gost/utils/exporter"
)

type ExporterConfig struct {
	TableName string
	exporter.BufferedExporterConfig
}

type Exporter[T any] struct {
	storage     *Storage
	bufExporter *exporter.BufferedExporter[T]
}

func NewExporter[T any](storage *Storage, cfg ExporterConfig) *Exporter[T] {
	e := &Exporter[T]{
		storage: storage,
	}
	e.bufExporter = exporter.NewBuffered[T](
		newExporterBatch[T](storage, cfg.TableName),
		cfg.BufferedExporterConfig,
	)
	return e
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
