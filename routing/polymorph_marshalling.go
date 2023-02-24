package routing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/bldsoft/gost/log"
	"go.mongodb.org/mongo-driver/bson"
)

const FieldNameType = "type"

type polymorphHeader struct {
	Type string `json:"type,omitempty" bson:"type,omitempty"`
}

// Helper for marshalling polymorphic objects
type PolymorphMarshaller[T comparable] struct {
	typeBijection[T, string]
}

func (pm *PolymorphMarshaller[T]) Register(name string, value T) {
	// if name == "" {
	// 	panic("routing: attempt to register empty name")
	// }

	if err := pm.Add(value, name); err != nil {
		panic(fmt.Sprintf("polymoprh marshaller: register %s", err))
	}

	log.Logger.TraceWithFields(log.Fields{"name": name, "type": reflect.TypeOf(value).String()}, "polymoprh marshaller: register polymorph object")
}

func (pm *PolymorphMarshaller[T]) AllocValue(name string) (val T, err error) {
	return pm.typeBijection.AllocValue(name)

}

func (pm *PolymorphMarshaller[T]) NameByValue(v T) (string, error) {
	name, ok := pm.GetObj(v)
	if !ok {
		return "", fmt.Errorf("polymoprh marshaller: name not registered for interface: %T", v)
	}
	return name, nil
}

func (pm *PolymorphMarshaller[T]) MarshalJSON(v T) ([]byte, error) {
	name, err := pm.NameByValue(v)
	if err != nil {
		return nil, err
	}
	return marshalJsonAndJoin(struct {
		Name string `json:"type"`
	}{
		Name: name,
	}, v)
}

func (pm *PolymorphMarshaller[T]) MarshalBSON(v T) ([]byte, error) {
	name, err := pm.NameByValue(v)
	if err != nil {
		return nil, err
	}

	return addTypeAndMashalBson(v, name)
}

func (pm *PolymorphMarshaller[T]) UnmarshalJSON(data []byte) (unmarshalled T, err error) {
	var temp polymorphHeader
	err = json.Unmarshal(data, &temp)
	if err != nil {
		return
	}

	ptr, res, err := pm.allocValue(temp.Type)
	if err != nil {
		return
	}
	ptri := ptr.Interface()

	err = json.Unmarshal(data, &ptri)
	if err != nil {
		return
	}

	return res.Interface().(T), nil
}

func (pm *PolymorphMarshaller[T]) UnmarshalBSON(data []byte) (unmarshalled T, err error) {
	var raw bson.Raw
	err = bson.Unmarshal(data, &raw)
	if err != nil {
		return
	}

	name, _ := raw.Lookup(FieldNameType).StringValueOK()

	ptr, res, err := pm.allocValue(name)
	if err != nil {
		return
	}
	ptri := ptr.Interface()

	if err = bson.Unmarshal(data, ptri); err != nil {
		return
	}

	return res.Interface().(T), nil
}

func marshalJsonAndJoin(objects ...interface{}) ([]byte, error) {
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

func toBsonMap(e interface{}) (m bson.M, err error) {
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

func addTypeAndMashalBson[T any](v T, typeName string) ([]byte, error) {
	if typeName != "" {
		//TODO: find more elegant way to do this
		m, err := toBsonMap(v)
		if err != nil {
			return nil, fmt.Errorf("polymorph marshaller: %w", err)
		}

		m[FieldNameType] = typeName
		return bson.Marshal(m)
	}

	return bson.Marshal(v)
}
