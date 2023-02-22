package routing

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/bldsoft/gost/utils"
)

// type to object bijection
type typeBijection[I comparable, V comparable] struct {
	objToConcreteType sync.Map // map[V]reflect.Type
	concreteTypeToObj sync.Map // map[reflect.Type]V
}

func (b *typeBijection[I, V]) Add(valueExample I, obj V) error {
	valueType := reflect.TypeOf(valueExample)
	if t, dup := b.objToConcreteType.LoadOrStore(obj, valueType); dup && t != valueType {
		return fmt.Errorf("duplicated type for %v: %q != %q", obj, t, valueType)
	}

	if o, dup := b.concreteTypeToObj.LoadOrStore(valueType, obj); dup && o != obj {
		return fmt.Errorf("duplicated obj for type %s: %v != %v", valueType, o, obj)
	}

	return nil
}

func (b *typeBijection[I, V]) GetObj(valueExample I) (v V, ok bool) {
	if obj, ok := b.concreteTypeToObj.Load(reflect.TypeOf(valueExample)); ok {
		return obj.(V), true
	}
	return
}

func (m *typeBijection[I, V]) GetType(obj V) (t reflect.Type, ok bool) {
	if typ, ok := m.objToConcreteType.Load(obj); ok {
		return typ.(reflect.Type), true
	}
	return
}

func (b *typeBijection[I, V]) AllocValue(obj V) (val I, err error) {
	_, v, err := b.allocValue(obj)
	if err != nil {
		return val, err
	}
	return v.Interface().(I), nil
}

func (b *typeBijection[I, V]) allocValue(obj V) (ptr, val reflect.Value, err error) {
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
