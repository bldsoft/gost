package routing

import (
	"fmt"
	"net"
	"reflect"
	"sort"
	"sync"

	"github.com/bldsoft/gost/auth/acl"
)

func init() {
	RegisterRule[*Rule]("")
	RegisterRule[RuleList]("multi")

	RegisterAction[MultiAction]("multi")
	RegisterAction[ActionRedirect]("redirect")
	RegisterAction[ActionModifyHeader]("modHeader")

	// field type
	RegisterCondition[MultiCondition]("multi")
	RegisterCondition[*FieldCondition[string]]("string")
	RegisterCondition[*FieldCondition[int]]("int")
	RegisterCondition[*FieldCondition[net.IP]]("ip")
	RegisterCondition[*FieldCondition[*string]]("*string")

	// exctractor | field type
	RegisterValueExtractor[IpExtractor, net.IP]("clientIP", "Client IP")
	RegisterValueExtractor[HostExtractor, string]("host", "Host")
	RegisterValueExtractor[PathExtractor, string]("path", "Path")
	RegisterValueExtractor[FileNameExtractor, string]("filename", "File name")
	RegisterValueExtractor[FileExtExtractor, string]("ext", "File extension")
	RegisterValueExtractor[*QueryExtractor, *string]("query", "Query param")
	RegisterValueExtractor[*HeaderExtractor, *string]("header", "Request Header")

	// matcher | field type
	RegisterValueMatcher[*MatcherClientIPAnyOf, net.IP]("anyOf", "Matches any of")
	RegisterValueMatcher[*MatcherClientIPNotAnyOf, net.IP]("notAnyOf", "Does not match any of")

	RegisterValueMatcher[*MatcherAnyOf[string], string]("anyOf", "Matches any of")
	RegisterValueMatcher[*MatcherNotAnyOf[string], string]("notAnyOf", "Does not match any of")

	RegisterValueMatcher[*MatcherAnyOfPtr[string], *string]("anyOf", "Matches any of")
	RegisterValueMatcher[*MatcherNotAnyOfPtr[string], *string]("notAnyOf", "Does not match any of")
	RegisterValueMatcher[*MatcherZero[*string], *string]("exists", "Exists")
	RegisterValueMatcher[*MatcherNotZero[*string], *string]("notExists", "Does not exists")

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

var fieldNameToType objToTypeMap[string, interface{}]                      // reflect.Type 		 -> string
var fieldTypeToExtractorMarshaller typeBijection[interface{}, interface{}] // reflect.Type 		<-> ValueExtractor[T]
var fieldNameToExtractor typeBijection[interface{}, string]                // ValueExtractor[T] <-> string
var fieldNameToExtractorDescription sync.Map                               // string -> FieldExtractorDescription

func RegisterValueExtractor[E ValueExtractor[T], T any](name string, label string) {
	var fieldValue T
	if err := fieldNameToType.Add(name, fieldValue); err != nil {
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

	extractorDescription := FieldExtractorDescription{
		Name:  name,
		Label: label,
	}
	extractorDescription.ArgDescription, err = getArgsDescription(extractorValue)
	if err != nil {
		panic(fmt.Sprintf("routing: extractor: %s", err))
	}

	if _, dup := fieldNameToExtractorDescription.LoadOrStore(name, extractorDescription); dup {
		panic(fmt.Sprintf("routing: duplicated extractor description for %s: ", name))
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

type FieldExtractorDescription struct {
	Name           string
	Label          string
	ArgDescription []ArgDescription
}

func ExtractorDescriptionsByFieldName(fieldName string) (*FieldExtractorDescription, error) {
	description, ok := fieldNameToExtractorDescription.Load(fieldName)
	if !ok {
		return nil, fmt.Errorf("routing: no extractor for %s", fieldName)
	}
	d := description.(FieldExtractorDescription)
	return &d, nil
}

// ===================================================

type ArgType string

const (
	ArgTypeInt         ArgType = "int"
	ArgTypeString      ArgType = "string"
	ArgTypeIntArray    ArgType = "[]int"
	ArgTypeStringArray ArgType = "[]string"
	ArgTypeIpRange     ArgType = "acl.IpRange"
)

func ValueToArgType(v interface{}) (ArgType, error) {
	switch any(v).(type) {
	case int:
		return ArgTypeInt, nil
	case string:
		return ArgTypeString, nil
	case []int:
		return ArgTypeIntArray, nil
	case []string:
		return ArgTypeStringArray, nil
	case acl.IpRange:
		return ArgTypeIpRange, nil
	}
	return "", fmt.Errorf("unsupported value type %T", v)
}

type MatcherDescription struct {
	Name           string
	Label          string
	ArgDescription []ArgDescription
}

type matcherTypeKeyType struct {
	fieldType   reflect.Type
	matcherName string
}

var matcherFieldTypeToFieldConditionType objToTypeMap[interface{}, Condition]   // reflect.Type -> *FieldCondition[T]
var matcherFieldTypeToMatcherMarshaller typeBijection[interface{}, interface{}] // reflect.Type <-> *PolymorphMarshaller[ValueMatcher[T]]
var matcherTypeBijection typeBijection[interface{}, matcherTypeKeyType]         // matcherTypeKeyType <-> ValueMatcher[T]
var fieldTypeToMatcherDescriptions sync.Map                                     // interface{} -> MatcherDescription

func matcherTypeKey[T any](matcherName string) matcherTypeKeyType {
	var fieldValue T
	return matcherTypeKeyType{
		fieldType:   reflect.TypeOf(fieldValue),
		matcherName: matcherName,
	}
}

func matcherMarshaller[T any]() (*PolymorphMarshaller[ValueMatcher[T]], error) {
	var fieldValue T
	valueMatcherMarshaller, ok := matcherFieldTypeToMatcherMarshaller.GetObj(fieldValue)
	if !ok {
		return nil, fmt.Errorf("no value matcher found for %T", fieldValue)
	}
	return valueMatcherMarshaller.(*PolymorphMarshaller[ValueMatcher[T]]), nil
}

func RegisterValueMatcher[M ValueMatcher[T], T any](name string, label string) {
	var fieldValue T
	valueMatcherMarshaller, err := matcherFieldTypeToMatcherMarshaller.AddOrGetObj(fieldValue, &PolymorphMarshaller[ValueMatcher[T]]{})
	if err != nil {
		panic(fmt.Sprintf("routing: %s", err))
	}

	var matcherValue M
	valueMatcherMarshaller.(*PolymorphMarshaller[ValueMatcher[T]]).Register(name, matcherValue)

	matcherDescription := MatcherDescription{
		Name:  name,
		Label: label,
	}
	matcherDescription.ArgDescription, err = getArgsDescription(matcherValue)
	if err != nil {
		panic(fmt.Sprintf("routing: matcher: %s", err))
	}

	fieldType := reflect.TypeOf(fieldValue)
	matchers, dup := fieldTypeToMatcherDescriptions.LoadOrStore(fieldType, &[]MatcherDescription{matcherDescription})
	if dup {
		appended := append(*matchers.(*[]MatcherDescription), matcherDescription)
		for !fieldTypeToMatcherDescriptions.CompareAndSwap(fieldType, matchers, &appended) {
			matchers, _ = fieldTypeToMatcherDescriptions.Load(fieldType)
		}
	}

	if err := matcherTypeBijection.Add(matcherValue, matcherTypeKey[T](name)); err != nil {
		panic(fmt.Sprintf("routing: %s", err))
	}

	var fieldCondtition FieldCondition[T]
	if err := matcherFieldTypeToFieldConditionType.Add(fieldType, fieldCondtition); err != nil {
		panic(fmt.Sprintf("routing: %s", err))
	}
}

// =======================================================

type FieldConditionDescription struct {
	Field              string
	FieldExtractorArgs []ArgDescription
	Matcher            string
	MatcherArgs        []ArgDescription
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
	extractorArgsDescription, err := getArgsDescription(extractor)
	if err != nil {
		return nil, err
	}

	matcherValue := condValue.FieldByName(fieldConditionMatcherFieldName).Elem()
	matcher := matcherValue.Interface()
	matcherTypeKey, ok := matcherTypeBijection.GetObj(matcher)
	if !ok {
		return nil, fmt.Errorf("failed to find matcher field for %T", matcher)
	}
	matcherArgsDescription, err := getArgsDescription(matcher)
	if err != nil {
		return nil, err
	}

	return &FieldConditionDescription{
		Field:              extractorName,
		FieldExtractorArgs: extractorArgsDescription,
		Matcher:            matcherTypeKey.matcherName,
		MatcherArgs:        matcherArgsDescription,
	}, nil
}

func FieldNames() []string {
	res := fieldNameToType.Keys()
	sort.Strings(res)
	return res
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

func BuildFieldCondition(field string, extractorArgs []Arg, op string, opArgs []Arg) (Condition, error) {
	fieldType, ok := fieldNameToType.GetType(field)
	if !ok {
		return nil, fmt.Errorf("routing: unknown field %s", field)
	}

	_, condValue, err := matcherFieldTypeToFieldConditionType.allocValue(fieldType)
	if err != nil {
		return nil, fmt.Errorf("routing: field condition: %w", err)
	}
	// condValue := reflect.ValueOf(cond).Addr()

	matcher, err := matcherTypeBijection.AllocValue(matcherTypeKeyType{
		fieldType:   fieldType,
		matcherName: op,
	})
	if err != nil {
		return nil, fmt.Errorf("routing: matcher: %w", err)
	}
	if err := setArgs(matcher, opArgs); err != nil {
		return nil, err
	}
	condValue.FieldByName(fieldConditionMatcherFieldName).Set(reflect.ValueOf(matcher))

	extractor, err := fieldNameToExtractor.AllocValue(field)
	if err != nil {
		return nil, fmt.Errorf("routing: extractor: %w", err)
	}
	if err := setArgs(extractor, extractorArgs); err != nil {
		return nil, err
	}
	condValue.FieldByName(fieldConditionExtractorFieldName).Set(reflect.ValueOf(extractor))

	return condValue.Addr().Interface().(Condition), nil
}

func FieldDecriptionByName(fieldName string) (*FieldExtractorDescription, []MatcherDescription, error) {
	extractorDescription, err := ExtractorDescriptionsByFieldName(fieldName)
	if err != nil {
		return nil, nil, err
	}
	matcherDescriptions, err := MatchersDescriptionsByFieldName(fieldName)
	if err != nil {
		return nil, nil, err
	}
	return extractorDescription, matcherDescriptions, nil
}
