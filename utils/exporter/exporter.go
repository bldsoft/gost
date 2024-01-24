package exporter

type Exporter[T any] interface {
	Export(items ...T) (n int, err error)
}

type Data[T any] interface {
	Send() error
	Len() int
	Add(...T) (n int, err error)
	Reset() error
}

type ExporterFunc[T any] struct {
	export func(items ...T) (n int, err error)
}

func Func[T any](export func(items ...T) (n int, err error)) ExporterFunc[T] {
	return ExporterFunc[T]{export}
}

func (e ExporterFunc[T]) Export(items ...T) (n int, err error) {
	return e.export(items...)
}
