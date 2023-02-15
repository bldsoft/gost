package routing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/bldsoft/gost/log"
	"go.mongodb.org/mongo-driver/bson"
)

const FieldNameConditionType = "type"

type polymorphHeader struct {
	Type string `json:"type,omitempty" bson:"type,omitempty"`
}

// Helper for marshalling polymorphic objects
type PolymorphMarshaller[T any] struct {
	nameToConcreteType sync.Map // map[string]reflect.Type
	concreteTypeToName sync.Map // map[reflect.Type]string
}

func (pm *PolymorphMarshaller[T]) Register(name string, value T) {
	// if name == "" {
	// 	panic("routing: attempt to register empty name")
	// }

	valueType := reflect.TypeOf(value)
	if t, dup := pm.nameToConcreteType.LoadOrStore(name, valueType); dup {
		panic(fmt.Sprintf("routing: registering duplicate condtition types for %q: %s != %s", name, t, valueType))
	}

	if n, dup := pm.concreteTypeToName.LoadOrStore(valueType, name); dup && n != name {
		panic(fmt.Sprintf("routing: registering duplicate names for %s: %q != %q", valueType, n, name))
	}

	log.Logger.TraceWithFields(log.Fields{"name": name, "type": valueType}, "register condition")
}

func (pm *PolymorphMarshaller[T]) AllocValue(name string) (val T, err error) {
	_, v, err := pm.allocValue(name)
	if err != nil {
		return
	}

	return v.Interface().(T), nil
}

// returns pointer and value of the regestered type (struct or pointer)
func (pm *PolymorphMarshaller[T]) allocValue(name string) (ptr, val reflect.Value, err error) {
	typi, ok := pm.nameToConcreteType.Load(string(name))
	if !ok {
		err = fmt.Errorf("routing: name not registered: %q", name)
		return
	}
	typ := typi.(reflect.Type)

	v := reflect.New(typ).Elem()
	if v.Type().Kind() == reflect.Ptr {
		v.Set(reflect.New(v.Type().Elem()))
		return v, v, nil
	}
	return v.Addr(), v, nil
}

func (pm *PolymorphMarshaller[T]) NameByValue(v T) (string, error) {
	valueType := reflect.TypeOf(v)
	name, ok := pm.concreteTypeToName.Load(valueType)
	if !ok {
		return "", fmt.Errorf("routing: name not registered for interface: %T", v)
	}
	return name.(string), nil
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
	valueType := reflect.TypeOf(v)
	name, ok := pm.concreteTypeToName.Load(valueType)
	if !ok {
		return nil, fmt.Errorf("routing: name not registered for interface: %q", valueType)
	}

	return marshalBsonAndJoin(struct {
		Name string `bson:"type"`
	}{
		Name: name.(string),
	}, v)
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

	name := raw.Lookup(FieldNameConditionType).StringValue()

	ptr, res, err := pm.allocValue(name)
	if err != nil {
		return
	}
	ptri := ptr.Interface()

	if err = bson.Unmarshal(data, &ptri); err != nil {
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

func marshalBsonAndJoin(lhs, rhs interface{}) ([]byte, error) {
	return bson.Marshal(struct {
		Lhs interface{} `bson:",inline"`
		Rhs interface{} `bson:",inline"`
	}{
		Lhs: lhs, Rhs: rhs,
	})
}
