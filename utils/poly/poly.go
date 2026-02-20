package poly

import (
	"encoding/json"
	"fmt"
	"reflect"

	json_utils "github.com/bldsoft/gost/utils/json"
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

func (f *Poly[T]) Type() string {
	typeString, ok := f.typesMap().GetObj(f.Value)
	if !ok {
		panic(fmt.Errorf("poly: unregistered type %T for interface %s", f.Value, f.interfaceString()))
	}
	return typeString
}

func (f Poly[T]) MarshalJSON() ([]byte, error) {
	return json_utils.MarshalJsonAndJoin(struct {
		Name string `json:"type"`
	}{
		Name: f.Type(),
	}, f.Value)
}

func (f *Poly[T]) UnmarshalJSON(data []byte) error {
	var temp polymorphHeader
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	ptr, res, err := f.typesMap().allocValue(temp.Type)
	if err != nil {
		panic(fmt.Errorf("poly: unregistered type %s for interface %s", temp.Type, f.interfaceString()))
	}
	ptri := ptr.Interface()

	if err = json.Unmarshal(data, &ptri); err != nil {
		return err
	}

	f.Value = res.Interface().(T)
	return nil
}

func (f Poly[T]) MarshalBSON() ([]byte, error) {
	return f.addTypeAndMashalBson(f.Value, f.Type())
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
