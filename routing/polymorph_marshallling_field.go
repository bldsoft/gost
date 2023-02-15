package routing

type marshallingField[T any] struct {
	Value               T
	polymorphMarshaller *PolymorphMarshaller[T]
}

func (f *marshallingField[T]) UnmarshalJSON(data []byte) error {
	var err error
	f.Value, err = f.polymorphMarshaller.UnmarshalJSON(data)
	return err
}

func (f *marshallingField[T]) UnmarshalBSON(data []byte) error {
	var err error
	f.Value, err = f.polymorphMarshaller.UnmarshalBSON(data)
	return err
}

func (f marshallingField[T]) MarshalJSON() ([]byte, error) {
	return f.polymorphMarshaller.MarshalJSON(f.Value)
}

func (f marshallingField[T]) MarshalBSON() ([]byte, error) {
	return f.polymorphMarshaller.MarshalBSON(f.Value)
}
