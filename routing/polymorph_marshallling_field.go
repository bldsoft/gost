package routing

import "reflect"

type marshallingField[T comparable] struct {
	Value               T
	polymorphMarshaller *PolymorphMarshaller[T]
}

func newMarshallingField[T comparable](m *PolymorphMarshaller[T], obj T) *marshallingField[T] {
	if v := reflect.ValueOf(obj); (v.Kind() == reflect.Ptr && v.IsNil()) || v.Kind() == reflect.Invalid {
		return nil
	}
	return &marshallingField[T]{Value: obj, polymorphMarshaller: m}
}

func makeMarshallingFields[T comparable](m *PolymorphMarshaller[T], obj ...T) []marshallingField[T] {
	res := make([]marshallingField[T], 0, len(obj))
	for _, o := range obj {
		res = append(res, marshallingField[T]{Value: o, polymorphMarshaller: m})
	}
	return res
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
