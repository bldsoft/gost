package config

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/viper"
)

type IConfig interface {
	SetDefaults()
	Validate() error
}

func ReadConfig(config IConfig, prefix string) {
	config.SetDefaults()
	ReadFromEnv(config, prefix)
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}
}

//FormatEnv formats all environment variables in yaml
func FormatEnv(config IConfig) string {
	d, _ := yaml.Marshal(config)
	return fmt.Sprintf("GOMAXPROCS: %d\n%s",
		runtime.GOMAXPROCS(0),
		string(d),
	)
}

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

		if tagValue == "" {
			tagValue = strings.ToUpper(configType.Field(i).Name)
		}
		if prefix != "" {
			tagToEnvKey = append(tagToEnvKey, tagValue, addPrefix(tagValue, prefix))
		}
		v.BindEnv(tagValue)
	}

	if len(tagToEnvKey) > 0 {
		v.SetEnvKeyReplacer(strings.NewReplacer(tagToEnvKey...))
	}
	return v.Unmarshal(&config)
}
