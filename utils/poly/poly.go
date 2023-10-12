package poly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

const FieldNameType = "type"

type polymorphHeader struct {
	Type string `json:"type,omitempty" bson:"type,omitempty"`
}

type Poly[T comparable] struct {
	Value T
}

func (f *Poly[T]) typesMap() *typeBijection[T, string] {
	interfaceType := reflect.TypeOf((*T)(nil)).Elem()
	typeNamesI, ok := interfaceTypeToTypesNames.Load(interfaceType)
	if !ok {
		panic(fmt.Errorf("poly: unregistered interface %s: register it with poly.Register", f.interfaceString()))
	}
	return typeNamesI.(*typeBijection[T, string])
}

func (f *Poly[T]) interfaceString() string {
	return fmt.Sprintln(reflect.TypeOf((*T)(nil)).Elem())
}

func (f *Poly[T]) typeString() string {
	typeString, ok := f.typesMap().GetObj(f.Value)
	if !ok {
		panic(fmt.Errorf("poly: unregistered type %T for interface %s", f.Value, f.interfaceString()))
	}
	return typeString
}

func (f Poly[T]) MarshalJSON() ([]byte, error) {
	return f.marshalJsonAndJoin(struct {
		Name string `json:"type"`
	}{
		Name: f.typeString(),
	}, f.Value)
}

func (f *Poly[T]) UnmarshalJSON(data []byte) error {
	var temp polymorphHeader
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	ptr, res, err := f.typesMap().allocValue(temp.Type)
	if err != nil {
		return err
	}
	ptri := ptr.Interface()

	if err = json.Unmarshal(data, &ptri); err != nil {
		return err
	}

	f.Value = res.Interface().(T)
	return nil
}

func (f *Poly[T]) marshalJsonAndJoin(objects ...interface{}) ([]byte, error) {
	if len(objects) == 0 {
		return nil, nil
	}
	marshalled := make([][]byte, 0, len(objects))
	for _, obj := range objects {
		b, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		if len(b) <= 2 { // "{}"
			continue
		}
		marshalled = append(marshalled, b)
	}

	if len(marshalled) == 1 {
		return marshalled[0], nil
	}

	for i, b := range marshalled {
		if i < len(objects)-1 {
			b = b[:len(b)-1]
		}
		if i > 0 {
			b = b[1:]
		}
		marshalled[i] = b
	}
	return bytes.Join(marshalled, []byte{','}), nil
}

func (f Poly[T]) MarshalBSON() ([]byte, error) {
	return f.addTypeAndMashalBson(f.Value, f.typeString())
}

func (f *Poly[T]) UnmarshalBSON(data []byte) error {
	var raw bson.Raw

	if err := bson.Unmarshal(data, &raw); err != nil {
		return err
	}

	name, _ := raw.Lookup(FieldNameType).StringValueOK()

	ptr, res, err := f.typesMap().allocValue(name)
	if err != nil {
		return err
	}
	ptri := ptr.Interface()

	if err = bson.Unmarshal(data, ptri); err != nil {
		return err
	}

	f.Value = res.Interface().(T)
	return nil
}

func (f *Poly[T]) toBsonMap(e interface{}) (m bson.M, err error) {
	data, err := bson.Marshal(e)
	if err != nil {
		return nil, err
	}
	err = bson.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (f *Poly[T]) addTypeAndMashalBson(v T, typeName string) ([]byte, error) {
	if typeName != "" {
		//TODO: find more elegant way to do this
		m, err := f.toBsonMap(v)
		if err != nil {
			return nil, fmt.Errorf("poly: %w", err)
		}

		m[FieldNameType] = typeName
		return bson.Marshal(m)
	}

	return bson.Marshal(v)
}
