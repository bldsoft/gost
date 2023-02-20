package routing

import (
	"fmt"
	"reflect"
	"sync"
)

func init() {
	RegisterRule[*Rule]("")

	RegisterAction[ActionRedirect]("redirect")

	RegisterCondition[*FieldCondition[string]]("")

	RegisterValueExtractor[HostExtractor, string]("host")

	RegisterValueMatcher[MatcherAnyOf, string]("anyOf")

}

var ruleMarshaller = &PolymorphMarshaller[IRule]{}

func RegisterRule[T IRule](name string) {
	var value T
	ruleMarshaller.Register(name, value)
}

var actionMarshaller = &PolymorphMarshaller[Action]{}

func RegisterAction[T Action](name string) {
	var value T
	actionMarshaller.Register(name, value)
}

var conditionPolymorphMarshaller = &PolymorphMarshaller[Condition]{}

func RegisterCondition[T Condition](name string) {
	var value T
	conditionPolymorphMarshaller.Register(name, value)
}

var fieldNameToType sync.Map                     // map[string]reflect.Type
var fieldTypeToValueExtractorMarshaller sync.Map // map[reflect.Type]ValueExtractor[T]

func valueExtractorMarshaller[T any]() (*PolymorphMarshaller[ValueExtractor[T]], error) {
	var fieldValue T
	fieldType := reflect.TypeOf(fieldValue)
	valueExtractorMarshaller, ok := fieldTypeToValueExtractorMarshaller.Load(fieldType)
	if !ok {
		return nil, fmt.Errorf("no value extractor found for %v", fieldType)
	}
	return valueExtractorMarshaller.(*PolymorphMarshaller[ValueExtractor[T]]), nil
}

func RegisterValueExtractor[E ValueExtractor[T], T any](name string) {
	var fieldValue T
	fieldType := reflect.TypeOf(fieldValue)
	valueExtractorMarshaller, _ := fieldTypeToValueExtractorMarshaller.LoadOrStore(fieldType, &PolymorphMarshaller[E]{})

	var extractorValue E
	valueExtractorMarshaller.(*PolymorphMarshaller[E]).Register(name, extractorValue)

	if t, dup := fieldNameToType.LoadOrStore(name, fieldType); dup && t != fieldType {
		panic(fmt.Sprintf("routing: registering duplicate field name %s: %s != %s", name, t, fieldType))
	}
}

var fieldTypeToValueMatcherMarshaller sync.Map // map[reflect.Type]ValueMatcher[T]
var fieldTypeToValueMatchers sync.Map          // map[reflect.Type][]ValueMatcher[T]
var matcherTypeToMatcherName sync.Map          // map[ValueMatcher[T]]string

func valueMatcherMarshaller[T any]() (*PolymorphMarshaller[ValueMatcher[T]], error) {
	var fieldValue T
	fieldType := reflect.TypeOf(fieldValue)
	valueMatcherMarshaller, ok := fieldTypeToValueMatcherMarshaller.Load(fieldType)
	if !ok {
		return nil, fmt.Errorf("no value matcher found for %v", fieldType)
	}
	return valueMatcherMarshaller.(*PolymorphMarshaller[ValueMatcher[T]]), nil
}

func RegisterValueMatcher[M ValueMatcher[T], T any](name string) {
	var fieldValue T
	fieldType := reflect.TypeOf(fieldValue)
	valueMatcherMarshaller, _ := fieldTypeToValueMatcherMarshaller.LoadOrStore(fieldType, &PolymorphMarshaller[M]{})

	var matcherValue M
	valueMatcherMarshaller.(*PolymorphMarshaller[M]).Register(name, matcherValue)

	matchers, dup := fieldTypeToValueMatchers.LoadOrStore(fieldType, []ValueMatcher[T]{matcherValue})
	if dup {
		for !fieldTypeToValueMatchers.CompareAndSwap(fieldType, matchers, append(matchers.([]ValueMatcher[T]), matcherValue)) {
			matchers, _ = fieldTypeToValueMatchers.Load(fieldType)
		}
	}

	matcherType := reflect.TypeOf(matcherValue)
	if n, dup := matcherTypeToMatcherName.LoadOrStore(matcherType, name); dup && n != name {
		panic(fmt.Sprintf("routing: registering duplicate names for %s: %q != %q", matcherType, n, name))
	}
}

func MatchersByFieldName(fieldName string) ([]string, error) {
	t, ok := fieldNameToType.Load(fieldName)
	if !ok {
		return nil, fmt.Errorf("routing: field name %s not registered", fieldName)
	}

	matchersI, ok := fieldTypeToValueMatchers.Load(t)
	if !ok {
		return nil, fmt.Errorf("routing: no matchers for type %s", t)
	}

	matchersV := reflect.ValueOf(matchersI)
	matcherNames := make([]string, 0, matchersV.Len())
	for i := 0; i < matchersV.Len(); i++ {
		matcherType := matchersV.Index(i).Elem().Type()
		matcherName, _ := matcherTypeToMatcherName.Load(matcherType)
		matcherNames = append(matcherNames, matcherName.(string))
	}
	return matcherNames, nil
}
