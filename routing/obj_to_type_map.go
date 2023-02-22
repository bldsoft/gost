package routing

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/bldsoft/gost/utils"
)

type objToTypeMap[V comparable, I comparable] struct {
	objToConcreteType sync.Map // map[V]reflect.Type
}

func (b *objToTypeMap[V, I]) Add(obj V, valueExample I) error {
	valueType := reflect.TypeOf(valueExample)
	if t, dup := b.objToConcreteType.LoadOrStore(obj, valueType); dup && t != valueType {
		return fmt.Errorf("duplicated type for %v: %q != %q", obj, t, valueType)
	}
	return nil
}

func (m *objToTypeMap[V, I]) Keys() []V {
	var res []V
	m.objToConcreteType.Range(func(key, value interface{}) bool {
		res = append(res, key.(V))
		return true
	})
	return res
}

func (m *objToTypeMap[V, I]) GetType(obj V) (t reflect.Type, ok bool) {
	if typ, ok := m.objToConcreteType.Load(obj); ok {
		return typ.(reflect.Type), true
	}
	return
}

func (b *objToTypeMap[V, I]) AllocValue(obj V) (val I, err error) {
	_, v, err := b.allocValue(obj)
	if err != nil {
		return val, err
	}
	return v.Interface().(I), nil
}

func (b *objToTypeMap[V, I]) allocValue(obj V) (ptr, val reflect.Value, err error) {
	typ, ok := b.GetType(obj)
	if !ok {
		return ptr, val, utils.ErrObjectNotFound
	}

	v := reflect.New(typ).Elem()
	if v.Type().Kind() == reflect.Ptr {
		v.Set(reflect.New(v.Type().Elem()))
		return v, v, nil
	}
	return v.Addr(), v, nil
}
