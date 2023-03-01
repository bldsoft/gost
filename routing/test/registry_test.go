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
		extractorDescr, matchersDescriptions, err := routing.FieldDecriptionByName(field)
		assert.NoError(t, err)
		assert.NotEmpty(t, extractorDescr)
		assert.NotEmpty(t, matchersDescriptions)

		extractorArgs := make([]routing.Arg, 0, len(extractorDescr.ArgDescription))
		for _, arg := range extractorDescr.ArgDescription {
			extractorArgs = append(extractorArgs, arg.Arg)
		}

		for _, matcherDescription := range matchersDescriptions {
			var cond routing.Condition
			var err error

			matcherArgs := make([]routing.Arg, 0, len(matcherDescription.ArgDescription))
			for _, arg := range matcherDescription.ArgDescription {
				matcherArgs = append(matcherArgs, arg.Arg)
			}

			cond, err = routing.BuildFieldCondition(field, extractorArgs, matcherDescription.Name, matcherArgs)
			assert.NoError(t, err)

			condDesctiption, err := routing.GetFieldConditionDescription(cond)
			assert.NoError(t, err)
			assert.Equal(t, field, condDesctiption.Field)
			assert.Equal(t, matcherDescription.ArgDescription, condDesctiption.MatcherArgs)
			assert.Equal(t, extractorDescr.ArgDescription, condDesctiption.FieldExtractorArgs)
		}
	}
}

func TestBuildFieldConditionAnyOf(t *testing.T) {
	cond, err := routing.BuildFieldCondition("host", nil, "anyOf", []routing.Arg{
		{
			Name:  "Values",
			Value: []string{"123"},
		},
	})
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
	assert.Equal(t, []string{"123"}, d.MatcherArgs[0].Value)

}
