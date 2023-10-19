package poly

import (
	"reflect"
)

// type to object bijection
type typeBijection[I comparable, V comparable] struct {
	concreteTypeToObj typeToObjMap[I, V]
	objToConcreteType objToTypeMap[V, I]
}

func (b *typeBijection[I, V]) AddOrGetObj(valueExample I, obj V) (actual V, err error) {
	if err := b.objToConcreteType.Add(obj, valueExample); err != nil {
		return actual, err
	}
	return b.concreteTypeToObj.AddOrGetObj(valueExample, obj)
}

func (b *typeBijection[I, V]) AllObj() []V {
	return b.objToConcreteType.Keys()
}

func (b *typeBijection[I, V]) Add(valueExample I, obj V) error {
	if err := b.objToConcreteType.Add(obj, valueExample); err != nil {
		return err
	}
	if err := b.concreteTypeToObj.Add(valueExample, obj); err != nil {
		return err
	}
	return nil
}

func (b *typeBijection[I, V]) GetObj(valueExample I) (v V, ok bool) {
	return b.concreteTypeToObj.GetObj(valueExample)
}

func (b *typeBijection[I, V]) GetType(obj V) (t reflect.Type, ok bool) {
	return b.objToConcreteType.GetType(obj)
}

func (b *typeBijection[I, V]) AllocValue(obj V) (val I, err error) {
	return b.objToConcreteType.AllocValue(obj)
}

func (b *typeBijection[I, V]) allocValue(obj V) (ptr, val reflect.Value, err error) {
	return b.objToConcreteType.allocValue(obj)
}
