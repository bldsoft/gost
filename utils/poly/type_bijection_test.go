package poly

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bldsoft/gost/utils"
)

func TestTypeBijectionCollision(t *testing.T) {
	t.Run("collisions", func(t *testing.T) {
		var b typeBijection[any, string]
		assert.NoError(t, b.Add(0, "int"))
		assert.Error(t, b.Add(0, "INT"))
		assert.Error(t, b.Add(false, "int"))
	})

	t.Run("double add", func(t *testing.T) {
		var b typeBijection[any, string]
		assert.NoError(t, b.Add(0, "int"))
		assert.Error(t, b.Add(0, "int"))
		assert.Error(t, b.Add(0, "int"))
	})

	t.Run("get obj", func(t *testing.T) {
		var b typeBijection[any, string]
		assert.NoError(t, b.Add(0, "int"))
		s, ok := b.GetObj(0)
		assert.True(t, ok)
		assert.Equal(t, "int", s)

		assert.NoError(t, b.Add("", "string"))
		s, ok = b.GetObj("")
		assert.True(t, ok)
		assert.Equal(t, "string", s)
	})

	t.Run("alloc value", func(t *testing.T) {
		var b typeBijection[any, string]
		_, err := b.AllocValue("not existing")
		assert.ErrorIs(t, utils.ErrObjectNotFound, err)

		assert.NoError(t, b.Add("", "string"))
		s, _ := b.AllocValue("string")
		assert.Equal(t, s, "")

		assert.NoError(t, b.Add(0, "int"))
		i, _ := b.AllocValue("int")
		assert.Equal(t, i, 0)

		var iptr *int
		assert.NoError(t, b.Add(iptr, "*int"))
		ptri, _ := b.AllocValue("*int")
		assert.Equal(t, *ptri.(*int), 0)

	})
}
