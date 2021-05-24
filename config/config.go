package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

func addPrefix(name string, prefix string) string {
	if name != "" && prefix != "" {
		return fmt.Sprintf("%s_%s", prefix, name)
	}
	return prefix + name
}

// ReadFromEnv ...
func ReadFromEnv(config interface{}, prefix string) error {
	v := viper.NewWithOptions(viper.KeyDelimiter("."))
	v.AutomaticEnv()
	tagToEnvKey := make([]string, 0)
	value := reflect.ValueOf(config)
	configType := reflect.Indirect(value).Type()

	if value.Kind() != reflect.Ptr || configType.Kind() != reflect.Struct {
		return errors.New("Not struct ptr")
	}

	for i := 0; i < configType.NumField(); i++ {
		tagValue := configType.Field(i).Tag.Get("mapstructure")
		if field := reflect.Indirect(reflect.Indirect(value).Field(i)); field.Kind() == reflect.Struct {
			if err := ReadFromEnv(field.Addr().Interface(), addPrefix(tagValue, prefix)); err != nil {
				return err
			}
			continue
		}

		if tagValue != "" {
			if prefix != "" {
				tagToEnvKey = append(tagToEnvKey, tagValue, addPrefix(tagValue, prefix))
			}
			v.BindEnv(tagValue)
		}
	}

	if len(tagToEnvKey) > 0 {
		v.SetEnvKeyReplacer(strings.NewReplacer(tagToEnvKey...))
	}
	return v.Unmarshal(&config)
}
