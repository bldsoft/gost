package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/bldsoft/gost/routing"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestRuleMarshal(t *testing.T) {
	tests := []struct {
		name string
		rule routing.IRule
	}{
		{
			name: "rule",
			rule: routing.NewRule(routing.NewFieldCondition(routing.Host, routing.MatchesAnyOf("example.com")), routing.ActionRedirect{Code: http.StatusMovedPermanently, Host: "google.com"}),
		},
	}

	for testNameSuffix, format := range map[string]struct {
		marshal   func(v any) ([]byte, error)
		unmarshal func(data []byte, v any) error
	}{
		"json": {json.Marshal, json.Unmarshal},
		"bson": {bson.Marshal, bson.Unmarshal},
	} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s_%s", tt.name, testNameSuffix), func(t *testing.T) {
				data, err := format.marshal(tt.rule)
				assert.NoError(t, err)

				t.Log("marsahlled: ", string(data))
				var rule routing.Rule
				err = format.unmarshal(data, &rule)
				assert.NoError(t, err)
				assert.Equal(t, tt.rule, &rule)
			})
		}
	}
}

func TestRuleListMarshal(t *testing.T) {
	tests := []struct {
		name string
		rule routing.IRule
	}{
		{
			name: "rules",
			rule: routing.JoinRules(
				routing.NewRule(routing.NewFieldCondition(routing.Host, routing.MatchesAnyOf("example.com")), routing.ActionRedirect{Code: http.StatusMovedPermanently, Host: "google.com"}),
				routing.NewRule(routing.NewFieldCondition(routing.Host, routing.MatchesAnyOf("example2.com")), routing.ActionRedirect{Code: http.StatusNotFound, Host: "google2.com"}),
			),
		},
	}

	for testNameSuffix, format := range map[string]struct {
		marshal   func(v any) ([]byte, error)
		unmarshal func(data []byte, v any) error
	}{
		"json": {json.Marshal, json.Unmarshal},
		"bson": {bson.Marshal, bson.Unmarshal},
	} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s_%s", tt.name, testNameSuffix), func(t *testing.T) {
				data, err := format.marshal(tt.rule)
				assert.NoError(t, err)

				t.Log("marsahlled: ", string(data))
				var rules routing.RuleList
				err = format.unmarshal(data, &rules)
				assert.NoError(t, err)
				assert.Equal(t, tt.rule, rules)
			})
		}
	}
}
