package clickhouse

import (
	"context"
	"sync/atomic"

	"github.com/bldsoft/gost/utils/exporter"
)

type ExporterConfig struct {
	TableName string
	exporter.BufferedExporterConfig
}

type Exporter[T any] struct {
	storage     *Storage
	bufExporter *exporter.BufferedExporter[T]
	isReady     *atomic.Bool
}

func NewExporter[T any](storage *Storage, cfg ExporterConfig) *Exporter[T] {
	e := &Exporter[T]{
		storage: storage,
	}
	e.bufExporter = exporter.NewBuffered(
		newExporterBatch[T](storage, cfg.TableName),
		cfg.BufferedExporterConfig,
	)
	e.isReady = &storage.isReady
	return e
}

func (e *Exporter[T]) Export(items ...T) (n int, err error) {
	if !e.isReady.Load() {
		return 0, nil
	}
	return e.bufExporter.Export(items...)
}

func (e *Exporter[T]) Run() error {
	return e.bufExporter.Run()
}

func (e *Exporter[T]) Stop(ctx context.Context) error {
	return e.bufExporter.Stop(ctx)
}
