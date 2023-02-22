package routing

import (
	"testing"

	"github.com/bldsoft/gost/utils"
	"github.com/stretchr/testify/assert"
)

func TestTypeBijectionCollision(t *testing.T) {

	var b typeBijection[interface{}, string]

	t.Run("collisions", func(t *testing.T) {
		assert.NoError(t, b.Add(0, "int"))
		assert.Error(t, b.Add(0, "INT"))
		assert.Error(t, b.Add(false, "int"))
	})

	t.Run("double add", func(t *testing.T) {
		assert.NoError(t, b.Add(0, "int"))
		assert.NoError(t, b.Add(0, "int"))
		assert.NoError(t, b.Add(0, "int"))
	})

	t.Run("get obj", func(t *testing.T) {
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
		s, err := b.AllocValue("not existing")
		assert.ErrorIs(t, utils.ErrObjectNotFound, err)

		assert.NoError(t, b.Add("", "string"))
		s, err = b.AllocValue("string")
		assert.Equal(t, s, "")

		assert.NoError(t, b.Add(0, "int"))
		i, err := b.AllocValue("int")
		assert.Equal(t, i, 0)

		var iptr *int
		assert.NoError(t, b.Add(iptr, "*int"))
		ptri, err := b.AllocValue("*int")
		assert.Equal(t, *ptri.(*int), 0)

	})

}
