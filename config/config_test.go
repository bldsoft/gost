package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type plainConfig struct {
	BoolValue   bool     `mapstructure:"BOOL_VALUE"`
	IntValue    int64    `mapstructure:"INT_VALUE"`
	StringValue string   `mapstructure:"STRING_VALUE"`
	SliceValue  []string `mapstructure:"SLICE_VALUE"`
}

func TestReadingPlainConfig(t *testing.T) {
	config := plainConfig{
		BoolValue:   false,
		IntValue:    123,
		StringValue: "123",
		SliceValue:  []string{"1", "2", "3"},
	}
	expected := plainConfig{
		BoolValue:   true,
		IntValue:    456,
		StringValue: "456",
		SliceValue:  []string{"4", "5", "6"},
	}

	configCopy := config
	assert.NoError(t, ReadFromEnv(&config))
	assert.True(t, reflect.DeepEqual(config, configCopy))

	os.Setenv("BOOL_VALUE", strconv.FormatBool(expected.BoolValue))
	os.Setenv("INT_VALUE", strconv.FormatInt(expected.IntValue, 10))
	os.Setenv("STRING_VALUE", expected.StringValue)
	os.Setenv("SLICE_VALUE", strings.Join(expected.SliceValue, ","))

	assert.NoError(t, ReadFromEnv(&config))
	assert.True(t, reflect.DeepEqual(config, expected))
	os.Clearenv()
}

type nestedConfig struct {
	StringValue string `mapstructure:"STRING_VALUE"`
}

type complexConfig struct {
	FirstConf  nestedConfig
	SecondConf nestedConfig  `mapstructure:"SECOND"`
	ThirdConf  nestedConfig  `mapstructure:"THIRD"`
	FourthConf *nestedConfig `mapstructure:"FOURTH"`
}

func newComplexConfig(value string) *complexConfig {
	return &complexConfig{
		FirstConf:  nestedConfig{StringValue: "first" + value},
		SecondConf: nestedConfig{StringValue: "second" + value},
		ThirdConf:  nestedConfig{StringValue: "third" + value},
		FourthConf: &nestedConfig{StringValue: "fourth" + value},
	}
}

func setToEnv(value *complexConfig, prefix string) {
	os.Setenv(prefix+"STRING_VALUE", value.FirstConf.StringValue)
	os.Setenv(prefix+"SECOND_STRING_VALUE", value.SecondConf.StringValue)
	os.Setenv(prefix+"THIRD_STRING_VALUE", value.ThirdConf.StringValue)
	os.Setenv(prefix+"FOURTH_STRING_VALUE", value.FourthConf.StringValue)
}

func TestReadingConfigComposition(t *testing.T) {
	config := newComplexConfig("123")
	expected := newComplexConfig("456")

	setToEnv(expected, "")

	assert.NoError(t, ReadFromEnv(config))
	assert.True(t, reflect.DeepEqual(config, expected))
	os.Clearenv()
}

func TestReadingManyNestedLayers(t *testing.T) {
	config := struct {
		Nested  *complexConfig
		NestedA *complexConfig `mapstructure:"A"`
		NestedB *complexConfig `mapstructure:"B"`
	}{
		Nested:  newComplexConfig("123"),
		NestedA: newComplexConfig("123"),
		NestedB: newComplexConfig("123"),
	}

	expected := config
	expected.Nested = newComplexConfig("456")
	expected.NestedA = newComplexConfig("456")
	expected.NestedB = newComplexConfig("456")

	setToEnv(expected.Nested, "")
	setToEnv(expected.NestedA, "A_")
	setToEnv(expected.NestedB, "B_")

	assert.NoError(t, ReadFromEnv(&config))
	assert.True(t, reflect.DeepEqual(config, expected))
	os.Clearenv()
}

func TestUntaggedField(t *testing.T) {
	config := struct {
		Untagged string
	}{
		Untagged: "123",
	}

	assert.NoError(t, ReadFromEnv(&config))
	assert.Equal(t, config.Untagged, "123")
}

func TestWrongTypes(t *testing.T) {
	config := struct{}{}
	var i int

	assert.Error(t, ReadFromEnv(config))
	assert.Error(t, ReadFromEnv(&i))
}
