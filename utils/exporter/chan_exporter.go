package exporter

type ChanExporter[T any] struct {
	ch chan<- T
}

func Chan[T any](ch chan<- T) *ChanExporter[T] {
	return &ChanExporter[T]{ch: ch}
}

func (e *ChanExporter[T]) Export(items ...T) (n int, err error) {
	for i, item := range items {
		select {
		case e.ch <- item:
		default:
			return i, nil
		}
	}
	return len(items), nil
}
