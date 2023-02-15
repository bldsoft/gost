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
		rule *routing.Rule
	}{
		{
			name: "rule",
			rule: routing.NewRule(routing.NewFieldCondition(routing.Host, routing.MatchesAnyOf("example.com")), routing.ActionRedirect{Code: http.StatusMovedPermanently, Host: "google.com"}),
		},
	}

	for testNameSuffix, _ := range map[string]struct {
		marshal   func(v any) ([]byte, error)
		unmarshal func(data []byte, v any) error
	}{
		"json": {json.Marshal, json.Unmarshal},
		"bson": {bson.Marshal, bson.Unmarshal},
	} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s_%s", tt.name, testNameSuffix), func(t *testing.T) {
				data, err := json.Marshal(tt.rule)
				assert.NoError(t, err)

				t.Log(string(data))
				var rule routing.Rule
				err = json.Unmarshal(data, &rule)
				assert.NoError(t, err)
				assert.Equal(t, tt.rule, &rule)
			})
		}
	}
}
