package exporter

// Exporter to Data adapter
type Slice[T any] struct {
	exporter Exporter[T]
	buf      []T
}

func NewSlice[T any](exporter Exporter[T]) *Slice[T] {
	return &Slice[T]{
		exporter: exporter,
	}
}

func (s *Slice[T]) WithBuf(buf []T) *Slice[T] {
	s.buf = buf
	return s
}

func (s *Slice[T]) Len() int {
	return len(s.buf)
}

func (s *Slice[T]) Send() error {
	_, err := s.exporter.Export(s.buf...)
	return err
}

func (s *Slice[T]) Add(items ...T) (n int, err error) {
	s.buf = append(s.buf, items...)
	return len(items), nil
}

func (s *Slice[T]) Reset() error {
	s.buf = s.buf[:0]
	return nil
}
