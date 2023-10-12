package poly

import (
	"fmt"
	"reflect"
	"sync"
)

type typeToObjMap[I comparable, V comparable] struct {
	concreteTypeToObj sync.Map // map[reflect.Type]V
}

func (b *typeToObjMap[I, V]) AddOrGetObj(valueExample I, obj V) (actual V, err error) {
	o, _ := b.concreteTypeToObj.LoadOrStore(reflect.TypeOf(valueExample), obj)
	return o.(V), nil
}

func (b *typeToObjMap[I, V]) Add(valueExample I, obj V) error {
	valueType := reflect.TypeOf(valueExample)
	if o, dup := b.concreteTypeToObj.LoadOrStore(valueType, obj); dup {
		return fmt.Errorf("duplicated obj for type %s: %v != %v", valueType, o, obj)
	}

	return nil
}

func (b *typeToObjMap[I, V]) GetObj(valueExample I) (v V, ok bool) {
	if obj, ok := b.concreteTypeToObj.Load(reflect.TypeOf(valueExample)); ok {
		return obj.(V), true
	}
	return
}
