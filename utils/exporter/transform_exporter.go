package exporter

type TransformExporter[T any, U any] struct {
	exporter Exporter[U]
	tranform func(T) U
}

func Transform[T, U any](exporter Exporter[U], transform func(T) U) *TransformExporter[T, U] {
	return &TransformExporter[T, U]{exporter, transform}
}

func (e *TransformExporter[T, U]) Export(items ...T) (n int, err error) {
	exportItems := make([]U, 0, len(items))
	for _, item := range items {
		exportItems = append(exportItems, e.tranform(item))
	}
	return e.exporter.Export(exportItems...)
}
