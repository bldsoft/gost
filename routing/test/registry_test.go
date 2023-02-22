package test

import (
	"net/http/httptest"
	"testing"

	"github.com/bldsoft/gost/routing"
	"github.com/stretchr/testify/assert"
)

func TestFieldNames(t *testing.T) {
	fields := routing.FieldNames()
	assert.NotEmpty(t, fields)
	assert.Contains(t, fields, "host")
}

func TestMatchersDescriptionsByFieldName(t *testing.T) {
	for _, field := range routing.FieldNames() {
		descriptions, err := routing.MatchersDescriptionsByFieldName(field)
		assert.NoError(t, err)
		assert.NotEmpty(t, descriptions)
	}
}

func TestBuildFieldCondition(t *testing.T) {
	for _, field := range routing.FieldNames() {
		descriptions, err := routing.MatchersDescriptionsByFieldName(field)
		assert.NoError(t, err)
		assert.NotEmpty(t, descriptions)
		for _, matcherDescription := range descriptions {
			var cond routing.Condition
			var err error
			var argExample interface{}
			switch matcherDescription.ArgType {
			case routing.ArgTypeInt:
				argExample = 0
				cond, err = routing.BuildFieldCondition(field, matcherDescription.Name, argExample.(int))
			case routing.ArgTypeString:
				argExample = ""
				cond, err = routing.BuildFieldCondition(field, matcherDescription.Name, argExample.(string))
			case routing.ArgTypeIntArray:
				argExample = []int{0}
				cond, err = routing.BuildFieldCondition(field, matcherDescription.Name, argExample.([]int))
			case routing.ArgTypeStringArray:
				argExample = []string{""}
				cond, err = routing.BuildFieldCondition(field, matcherDescription.Name, argExample.([]string))
			default:
				assert.Fail(t, "unknown arg type")
			}
			assert.NoError(t, err)

			condDesctiption, err := routing.GetFieldConditionDescription(cond)
			assert.NoError(t, err)
			assert.Equal(t, field, condDesctiption.Field)
			assert.Equal(t, matcherDescription.Name, condDesctiption.Matcher)
			assert.Equal(t, argExample, condDesctiption.Args)
		}
	}
}

func TestBuildFieldConditionAnyOf(t *testing.T) {
	matcherDescriptions, err := routing.MatchersDescriptionsByFieldName("host")
	t.Log(matcherDescriptions)
	assert.NoError(t, err)
	assert.NotEmpty(t, matcherDescriptions)

	for _, d := range matcherDescriptions {
		cond, err := routing.BuildFieldCondition("host", d.Name, []string{"123"})
		assert.NoError(t, err)
		assert.NotNil(t, cond)

		m, err := cond.Match(httptest.NewRequest("GET", "http://example.com", nil))
		assert.NoError(t, err)
		assert.False(t, m)

		m, err = cond.Match(httptest.NewRequest("GET", "http://123", nil))
		assert.NoError(t, err)
		assert.True(t, m)

		d, err := routing.GetFieldConditionDescription(cond)
		assert.NoError(t, err)
		assert.Equal(t, "anyOf", d.Matcher)
		assert.Equal(t, []string{"123"}, d.Args)

	}
}
