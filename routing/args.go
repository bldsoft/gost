package routing

import (
	"errors"
	"fmt"
	"reflect"
)

type Arg struct {
	Name  string
	Value interface{}
}

func (p Arg) ArgType() (ArgType, error) {
	return ValueToArgType(p.Value)
}

type ArgDescription struct {
	Arg
	Label       string
	Description string
}

func setArgs(obj interface{}, args []Arg) error {
	if len(args) == 0 {
		return nil
	}
	value := reflect.ValueOf(obj)
	if value.Kind() != reflect.Ptr && reflect.Indirect(value).Kind() == reflect.Struct {
		return errors.New("obj must be a pointer to struct")
	}

	for _, arg := range args {
		field := value.Elem().FieldByName(arg.Name)
		if !field.IsValid() {
			return fmt.Errorf("%s not found", arg.Name)
		}
		if !field.Type().AssignableTo(reflect.TypeOf(arg.Value)) {
			return fmt.Errorf("wrong type for %s: %s != %s", arg.Name, reflect.TypeOf(arg.Value), field.Type())
		}
		field.Set(reflect.ValueOf(arg.Value))
	}

	return nil
}

const DescriptionTag = "descirption"
const LabelTag = "label"

func getArgsDescription(obj interface{}) ([]ArgDescription, error) {
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value = reflect.New(value.Type().Elem()).Elem()
		} else {
			value = reflect.Indirect(value)
		}
	}
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not a struct %v", obj)
	}

	var res []ArgDescription
	for i := 0; i < value.Type().NumField(); i++ {
		field := value.Type().Field(i)

		res = append(res, ArgDescription{
			Arg: Arg{
				Name:  field.Name,
				Value: value.Field(i).Interface(),
			},
			Description: field.Tag.Get(DescriptionTag),
			Label:       field.Tag.Get(LabelTag),
		})
	}
	return res, nil
}
