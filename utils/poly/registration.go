package poly

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/bldsoft/gost/log"
)

// must contain all concrete types for all interfaces so that polymorphic fields can be unmarshaled
var interfaceTypeToTypesNames sync.Map // reflict.Type -> *typeBijection[T, string]

type Registrator[Iface comparable] struct{}

func Register[Iface comparable]() Registrator[Iface] {
	return Registrator[Iface]{}
}

func (r Registrator[I]) Type(name string, value I) Registrator[I] {
	register[I](name, value)
	return r
}

func register[Iface comparable](name string, value Iface) {
	interfaceType := reflect.TypeOf((*Iface)(nil)).Elem()
	typesMap := &typeBijection[Iface, string]{}
	if actual, ok := interfaceTypeToTypesNames.LoadOrStore(interfaceType, typesMap); ok {
		typesMap = actual.(*typeBijection[Iface, string])
	}

	if err := typesMap.Add(value, name); err != nil {
		panic(fmt.Sprintf("polymoprh marshaller: register %s", err))
	}
	log.Logger.TraceWithFields(log.Fields{"name": name, "type": reflect.TypeOf(value).String()}, "poly: polymorphic object registered")
}
