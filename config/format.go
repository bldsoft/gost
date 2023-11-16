package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

const DescriptionTagName = "description"

type Formatter interface {
	WriteParam(envName, value string, descriptions []string) error
}

type paramEntry struct {
	Value       string
	Description []string
}

type paramList struct {
	envToDescription map[string]*paramEntry
	orderedKeys      []string
}

func newDescriptionList() *paramList {
	return &paramList{envToDescription: make(map[string]*paramEntry)}
}

func (l *paramList) Add(env, value, description string) {
	if entry, ok := l.envToDescription[env]; !ok {
		l.orderedKeys = append(l.orderedKeys, env)
		l.envToDescription[env] = &paramEntry{Description: []string{description}, Value: value}
	} else {
		entry.Description = append(entry.Description, description)
	}
}

func (l *paramList) WriteTo(formatter Formatter) error {
	for _, env := range l.orderedKeys {
		entry := l.envToDescription[env]
		if err := formatter.WriteParam(env, entry.Value, entry.Description); err != nil {
			return err
		}
	}
	return nil
}

func formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Slice:
		if v.Len() == 0 {
			return ""
		}
		var sb strings.Builder
		for i := 0; i < v.Len(); i++ {
			sb.WriteString(fmt.Sprintf("%v,", v.Index(i)))
		}
		return sb.String()[:sb.Len()-1]
	default:
		return fmt.Sprintf("%v", v)
	}
}

// WriteConfigDescription writes config description to formatter.
// It uses "description" tag to fill description column.
// If there config has fields with the same env name their descriptions are concateneted.
// if the field tag is "-", the field is always omitted.
func WriteConfigDescription(config interface{}, envPrefix string, formatter Formatter) error {
	list := newDescriptionList()
	if err := iterateFields(config, envPrefix, nil, func(envVarName, envNamePrefix string, field reflect.StructField, value reflect.Value) error {
		if !field.IsExported() {
			return nil
		}
		if description := field.Tag.Get(DescriptionTagName); description != "-" {
			list.Add(addPrefix(envVarName, envNamePrefix), formatValue(value), description)
		}
		return nil
	}, nil); err != nil {
		return err
	}

	return list.WriteTo(formatter)
}

func WriteMarkdownDescription(filename string, config interface{}, envPrefix string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return WriteConfigDescription(config, envPrefix, NewMarkdownFormatter(file, true))
}
