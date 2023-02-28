package routing

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"sync"
)

func init() {
	RegisterRule[*Rule]("")
	RegisterRule[RuleList]("multi")

	RegisterAction[MultiAction]("multi")
	RegisterAction[ActionRedirect]("redirect")

	// field type | matcher argument params
	RegisterCondition[MultiCondition]("multi")
	RegisterCondition[*FieldCondition[string, []string]]()
	RegisterCondition[*FieldCondition[int, []int]]()

	// exctractor | field type
	RegisterValueExtractor[HostExtractor, string]("host")
	RegisterValueExtractor[IpExtractor, net.IP]("clientIP")
	RegisterValueExtractor[PathExtractor, string]("path")
	RegisterValueExtractor[FileNameExtractor, string]("filename")
	RegisterValueExtractor[FileExtExtractor, string]("ext")
	RegisterValueExtractor[QueryExtractor, url.Values]("query")
	RegisterValueExtractor[HeaderExtractor, http.Header]("header")

	// matcher | field type | matcher argument params type
	RegisterValueMatcher[*MatcherAnyOf[string], string, []string]("anyOf", "Matches any of")

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

func RegisterCondition[T Condition](name ...string) {
	var value T
	if len(name) == 0 {
		name = append(name, reflect.TypeOf(value).String())
	}
	conditionPolymorphMarshaller.Register(name[0], value)
}

// =======================================================

var fieldNameToType typeBijection[interface{}, string]                     // reflect.Type 		<-> string
var fieldTypeToExtractorMarshaller typeBijection[interface{}, interface{}] // reflect.Type 		<-> ValueExtractor[T]
var fieldNameToExtractor typeBijection[interface{}, string]                // ValueExtractor[T] <-> string

func RegisterValueExtractor[E ValueExtractor[T], T any](name string) {
	var fieldValue T
	if err := fieldNameToType.Add(fieldValue, name); err != nil {
		panic(fmt.Sprintf("routing: %s", err))
	}

	extractorMarshallerI, err := fieldTypeToExtractorMarshaller.AddOrGetObj(fieldValue, &PolymorphMarshaller[ValueExtractor[T]]{})
	if err != nil {
		panic(fmt.Sprintf("routing: %s", err))
	}
	extractorMarshaller := extractorMarshallerI.(*PolymorphMarshaller[ValueExtractor[T]])

	var extractorValue E
	extractorMarshaller.Register(name, extractorValue)

	if fieldNameToExtractor.Add(extractorValue, name); err != nil {
		panic(fmt.Sprintf("routing: %s", err))
	}
}

func valueExtractorMarshaller[T any]() (*PolymorphMarshaller[ValueExtractor[T]], error) {
	var fieldValue T
	valueExtractorMarshaller, ok := fieldTypeToExtractorMarshaller.GetObj(fieldValue)
	if !ok {
		return nil, fmt.Errorf("no value extractor found for %T", fieldValue)
	}
	return valueExtractorMarshaller.(*PolymorphMarshaller[ValueExtractor[T]]), nil
}

// ===================================================

type ArgType string

const (
	ArgTypeInt         ArgType = "int"
	ArgTypeString      ArgType = "string"
	ArgTypeIntArray    ArgType = "[]int"
	ArgTypeStringArray ArgType = "[]string"
)

type MatcherDescription struct {
	Name    string
	Label   string
	ArgType ArgType
}

type marshallerGenericTypes struct {
	fieldType reflect.Type
	argsType  reflect.Type
}

type matcherTypeKeyType struct {
	marshallerGenericTypes
	matcherName string
}

var matcherInterfaceToFieldConditionType objToTypeMap[marshallerGenericTypes, Condition] // marshallerGenericTypes -> *FieldCondition[T,A]
var matcherInterfaceToMatcherMarshaller sync.Map                                         // marshallerGenericTypes -> *PolymorphMarshaller[ValueMatcher[T, A]]
var matcherTypeBijection typeBijection[interface{}, matcherTypeKeyType]                  // matcherTypeKeyType <-> ValueMatcher[T,A]
var fieldTypeToMatcherDescriptions sync.Map                                              // interface{} -> MatcherDescription

func matcherMarshallerKey[T, A any]() marshallerGenericTypes {
	var fieldValue T
	var argsValue A
	return matcherMarshallerKeyFromValues(fieldValue, argsValue)
}

func matcherMarshallerKeyFromValues[T, A any](fieldValue T, argsValue A) marshallerGenericTypes {
	fieldType := reflect.TypeOf(fieldValue)
	argsType := reflect.TypeOf(argsValue)
	return marshallerGenericTypes{fieldType, argsType}
}

func matcherTypeKey[T, A any](matcherName string) matcherTypeKeyType {
	var fieldValue T
	var argsValue A
	return matcherTypeKeyType{
		marshallerGenericTypes: marshallerGenericTypes{
			fieldType: reflect.TypeOf(fieldValue),
			argsType:  reflect.TypeOf(argsValue),
		},
		matcherName: matcherName,
	}
}

func matcherMarshaller[T, A any]() (*PolymorphMarshaller[ValueMatcher[T, A]], error) {
	key := matcherMarshallerKey[T, A]()
	valueMatcherMarshaller, ok := matcherInterfaceToMatcherMarshaller.Load(key)
	if !ok {
		return nil, fmt.Errorf("no value matcher found for %v", key)
	}
	return valueMatcherMarshaller.(*PolymorphMarshaller[ValueMatcher[T, A]]), nil
}

func RegisterValueMatcher[M ValueMatcher[T, A], T, A any](name string, label string) {
	key := matcherMarshallerKey[T, A]()
	valueMatcherMarshaller, _ := matcherInterfaceToMatcherMarshaller.LoadOrStore(key, &PolymorphMarshaller[ValueMatcher[T, A]]{})

	var matcherValue M
	valueMatcherMarshaller.(*PolymorphMarshaller[ValueMatcher[T, A]]).Register(name, matcherValue)

	matcherDescription := MatcherDescription{
		Name:    name,
		Label:   label,
		ArgType: ArgType(key.argsType.String()),
	}

	matchers, dup := fieldTypeToMatcherDescriptions.LoadOrStore(key.fieldType, &[]MatcherDescription{matcherDescription})
	if dup {
		appended := append(*matchers.(*[]MatcherDescription), matcherDescription)
		for !fieldTypeToMatcherDescriptions.CompareAndSwap(key.fieldType, matchers, &appended) {
			matchers, _ = fieldTypeToMatcherDescriptions.Load(key.fieldType)
		}
	}

	matcherKey := matcherTypeKey[T, A](name)
	if err := matcherTypeBijection.Add(matcherValue, matcherKey); err != nil {
		panic(fmt.Sprintf("routing: %s", err))
	}

	var fieldCondtition FieldCondition[T, A]
	if err := matcherInterfaceToFieldConditionType.Add(matcherKey.marshallerGenericTypes, fieldCondtition); err != nil {
		panic(fmt.Sprintf("routing: %s", err))
	}
}

// =======================================================

type FieldConditionDescription struct {
	Field   string
	Matcher string
	Args    interface{}
}

func GetFieldConditionDescription(fieldCondition Condition) (*FieldConditionDescription, error) {
	condValue := reflect.ValueOf(fieldCondition)
	if condValue.Kind() == reflect.Ptr {
		condValue = condValue.Elem()
	}
	extractor := condValue.FieldByName(fieldConditionExtractorFieldName).Interface()
	extractorName, ok := fieldNameToExtractor.GetObj(extractor)
	if !ok {
		return nil, fmt.Errorf("failed to find extractor field for %T", extractor)
	}

	matcherValue := condValue.FieldByName(fieldConditionMatcherFieldName).Elem()
	matcher := matcherValue.Interface()
	matcherTypeKey, ok := matcherTypeBijection.GetObj(matcher)
	if !ok {
		return nil, fmt.Errorf("failed to find matcher field for %T", matcher)
	}
	argValues := matcherValue.MethodByName(ArgsMethodName).Call(nil)

	return &FieldConditionDescription{
		Field:   extractorName,
		Matcher: matcherTypeKey.matcherName,
		Args:    argValues[0].Interface(),
	}, nil
}

func FieldNames() []string {
	return fieldNameToType.AllObj()
}

func MatchersDescriptionsByFieldName(fieldName string) ([]MatcherDescription, error) {
	t, ok := fieldNameToType.GetType(fieldName)
	if !ok {
		return nil, fmt.Errorf("routing: unknown field %s", fieldName)
	}

	matcherDescriptions, ok := fieldTypeToMatcherDescriptions.Load(t)
	if !ok {
		return nil, fmt.Errorf("routing: no matchers for type %s", t)
	}
	return *matcherDescriptions.(*[]MatcherDescription), nil
}

func BuildFieldCondition[A any](field, op string, args A) (Condition, error) {
	fieldType, ok := fieldNameToType.GetType(field)
	if !ok {
		return nil, fmt.Errorf("routing: unknown field %s", field)
	}

	argsType := reflect.TypeOf(args)
	key := matcherTypeKeyType{
		marshallerGenericTypes: marshallerGenericTypes{
			argsType:  reflect.TypeOf(args),
			fieldType: fieldType,
		},
		matcherName: op,
	}

	_, condValue, err := matcherInterfaceToFieldConditionType.allocValue(key.marshallerGenericTypes)
	if err != nil {
		return nil, fmt.Errorf("routing: field condition: %w", err)
	}
	// condValue := reflect.ValueOf(cond).Addr()

	matcher, err := matcherTypeBijection.AllocValue(key)
	if !ok {
		return nil, fmt.Errorf("routing: no matchers %s for type %s, %s", op, fieldType, argsType)
	}
	matcher.(IArgs[A]).SetArgs(args)
	condValue.FieldByName(fieldConditionMatcherFieldName).Set(reflect.ValueOf(matcher))

	extractor, err := fieldNameToExtractor.AllocValue(field)
	if err != nil {
		return nil, fmt.Errorf("routing: %w", err)
	}
	condValue.FieldByName(fieldConditionExtractorFieldName).Set(reflect.ValueOf(extractor))

	return condValue.Addr().Interface().(Condition), nil
}
