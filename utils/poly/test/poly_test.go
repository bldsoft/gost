package test

import (
	"encoding/json"
	"testing"

	"github.com/bldsoft/gost/utils/poly"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	poly.Register[SomeInterface]().
		Type("A", A{}).
		Type("B", B{})
}

type SomeInterface interface {
	SomeMethod()
}

type SomeOtherInterface interface {
	SomeOtherMethod()
}

type A struct {
	AField string
}

func (A) SomeMethod()      {}
func (A) SomeOtherMethod() {}

type B struct {
	BField string
}

func (B) SomeMethod()      {}
func (B) SomeOtherMethod() {}

type C struct {
	CField string
}

func (C) SomeMethod()      {}
func (C) SomeOtherMethod() {}

type Container struct {
	SomeField poly.Poly[SomeInterface]
}

func TestPolyJSONMarshal(t *testing.T) {

	testCases := []struct {
		value interface{}
		data  string
	}{
		{poly.Poly[SomeInterface]{A{AField: "AValue"}}, `{"type":"A","AField":"AValue"}`},
		{poly.Poly[SomeInterface]{B{BField: "BValue"}}, `{"type":"B","BField":"BValue"}`},
		{[]poly.Poly[SomeInterface]{{A{AField: "AValue"}}, {B{BField: "BValue"}}}, `[{"type":"A","AField":"AValue"},{"type":"B","BField":"BValue"}]`},
		{&Container{poly.Poly[SomeInterface]{A{AField: "AValue"}}}, `{"SomeField":{"type":"A","AField":"AValue"}}`},
		{&Container{poly.Poly[SomeInterface]{B{BField: "BValue"}}}, `{"SomeField":{"type":"B","BField":"BValue"}}`},
	}

	for _, test := range testCases {
		t.Run("marhsal", func(t *testing.T) {
			data, err := json.Marshal(test.value)
			require.NoError(t, err)
			require.Equal(t, test.data, string(data))
		})
	}

}

func testJSONUnmarshal[T any](t *testing.T, data string, expected T) {
	var value T
	err := json.Unmarshal([]byte(data), &value)
	require.NoError(t, err)
	require.Equal(t, expected, value)
}

func TestPolyJSONUnmarshal(t *testing.T) {
	testJSONUnmarshal[Container](t, `{"SomeField":{"type":"A","AField":"AValue"}}`, Container{poly.Poly[SomeInterface]{A{AField: "AValue"}}})
	testJSONUnmarshal[poly.Poly[SomeInterface]](t, `{"type":"A","AField":"AValue"}`, poly.Poly[SomeInterface]{A{AField: "AValue"}})
}

func testBSON[T any](t *testing.T, val T) {
	data, err := bson.Marshal(val)
	require.NoError(t, err)
	var new T
	err = bson.Unmarshal(data, &new)
	require.NoError(t, err)
	require.Equal(t, val, new)
}

func TestPolyBSON(t *testing.T) {
	testBSON(t, poly.Poly[SomeInterface]{Value: A{AField: "AValue"}})
	testBSON(t, Container{poly.Poly[SomeInterface]{A{AField: "AValue"}}})
}

func TestPolyUnregistered(t *testing.T) {
	t.Run("unmarshal not registered interface", func(t *testing.T) {
		var f poly.Poly[SomeOtherInterface]
		require.Panics(t, func() { _ = json.Unmarshal([]byte(`{}`), &f) })
	})

	t.Run("marshal not registered interface", func(t *testing.T) {
		var f poly.Poly[SomeOtherInterface]
		require.Panics(t, func() { _, _ = json.Marshal(f) })
	})

	t.Run("unmarshal not registered type", func(t *testing.T) {
		var container Container
		require.Panics(t, func() { _ = json.Unmarshal([]byte(`{"SomeField":{"type":"C","CField":"CValue"}}`), &container) })
	})

	t.Run("marshal not registered type", func(t *testing.T) {
		require.Panics(t, func() { _, _ = json.Marshal(&Container{poly.Poly[SomeInterface]{C{CField: "CValue"}}}) })
	})
}
