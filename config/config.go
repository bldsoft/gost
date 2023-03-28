package config

import (
	"errors"
	"fmt"
	"log"
	"os"
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

// ReadConfig reads config from the envrionment and validates it.
// config can include nested IConfig structs.
// ReadConfig calls SetDefaults of all IConfig structs, so you don't have to call SetDefaults of nested configs in parent SetDefualts.
// Do it only if you want to rewrite a child config defaults in the parent config.
// For the same reason you don't need call Validate of the nested IConfigs in the parent Validate method.
//
// See ReadFromEnv for detail of binding config fields with names of environment variables.
//
// Set CONFIG_DESCRIPTION_PATH env var to generate Markdown config description.
// See WriteConfigDescription for detail of the description generating.
func ReadConfig(config IConfig, envPrefix string) {
	if err := SetDefaults(config); err != nil {
		log.Fatalf("Failed to set config defaults: %v", err)
	}

	path := os.Getenv("CONFIG_DESCRIPTION_PATH")
	if len(path) > 0 {
		WriteMarkdownDescription(path, config, envPrefix)
	}

	if err := ReadFromEnv(config, envPrefix); err != nil {
		log.Fatalf("Failed read config from environment: %v", err)
	}

	if err := Validate(config); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}
}

// FormatEnv formats all environment variables in yaml
func FormatEnv(config IConfig) string {
	d, _ := yaml.Marshal(config)
	return fmt.Sprintf("GOMAXPROCS: %d\n%s",
		runtime.GOMAXPROCS(0),
		string(d),
	)
}

func foreachConfig(config interface{}, f func(cfg IConfig) error) error {
	return iterateFields(config, "", nil, nil, func(cfg interface{}) error {
		if c, ok := cfg.(IConfig); ok {
			return f(c)
		}
		return nil
	})
}

// SetDefaults sets default values of config and all nested IConfig structs
// SetDefaults method of nested IConfig is called before its parent. So you can rewrite default values of the child config in the parent if nessesary
func SetDefaults(config interface{}) error {
	return foreachConfig(config, func(cfg IConfig) error {
		cfg.SetDefaults()
		return nil
	})
}

// Validate validates config and all nested IConfig structs
func Validate(config interface{}) error {
	return foreachConfig(config, func(cfg IConfig) error {
		return cfg.Validate()
	})
}

func addPrefix(name string, prefix string) string {
	if name != "" && prefix != "" {
		return fmt.Sprintf("%s_%s", prefix, name)
	}
	return prefix + name
}

// ReadFromEnv reads config from environment variable
// The name of environment variable is set via "mapstructure" tag. If tag isn't set, the env name is the struct field name.
// Prefix for the env variable name consists of envPrefix and "mapstructure" values of all parent structs joined together with "_" (see example)
func ReadFromEnv(config interface{}, envPrefix string) error {
	v := viper.NewWithOptions(
		viper.KeyDelimiter("."),
	)
	v.AutomaticEnv()

	var tagToEnvKeyStack [][]string
	return iterateFields(config, envPrefix, func(cfg interface{}) error {
		tagToEnvKeyStack = append(tagToEnvKeyStack, nil) // push
		return nil
	}, func(envVarName, envNamePrefix string, _ reflect.StructField, _ reflect.Value) error {
		tagToEnvKeyStack[len(tagToEnvKeyStack)-1] = append(tagToEnvKeyStack[len(tagToEnvKeyStack)-1], envVarName, addPrefix(envVarName, envNamePrefix))
		return v.BindEnv(envVarName)
	}, func(cfg interface{}) error {
		top := tagToEnvKeyStack[len(tagToEnvKeyStack)-1]
		v.SetEnvKeyReplacer(strings.NewReplacer(top...))
		tagToEnvKeyStack = tagToEnvKeyStack[:len(tagToEnvKeyStack)-1] // pop
		return v.Unmarshal(&cfg)
	})
}

type fieldCallback func(envVarName string, envNamePrefix string, field reflect.StructField, value reflect.Value) error
type structCallback func(cfg interface{}) error

// iterateFields is a helper function for ReadFrom and WriteConfigDescription.
// The function traverses config and its nested structs using DFS algorythm, calling fieldCallback for each list field.
// Before processing a struct the function calls startSructCb, after - finishStructCb
// A "mapstructure" tag of struct appends envNamePrefix for all internal fields
func iterateFields(config interface{}, prefix string, startSructCb structCallback, fieldCb fieldCallback, finishStructCb structCallback) error {
	value := reflect.ValueOf(config)
	configType := reflect.Indirect(value).Type()

	if value.Kind() != reflect.Ptr || configType.Kind() != reflect.Struct {
		return errors.New("Not struct ptr")
	}

	if startSructCb != nil {
		if err := startSructCb(config); err != nil {
			return err
		}
	}

	for i := 0; i < configType.NumField(); i++ {
		field := configType.Field(i)
		if !field.IsExported() {
			continue
		}
		tagValue := field.Tag.Get("mapstructure")
		if field := reflect.Indirect(reflect.Indirect(value).Field(i)); field.Kind() == reflect.Struct {
			if err := iterateFields(field.Addr().Interface(), addPrefix(tagValue, prefix), startSructCb, fieldCb, finishStructCb); err != nil {
				return err
			}
			continue
		}

		if fieldCb == nil {
			continue
		}

		if tagValue == "" {
			tagValue = strings.ToUpper(field.Name)
		}

		if err := fieldCb(tagValue, prefix, field, value.Elem().Field(i)); err != nil {
			return err
		}

	}

	if finishStructCb != nil {
		return finishStructCb(config)
	}
	return nil
}
