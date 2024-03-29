package mongo

import (
	"testing"

	"github.com/bldsoft/gost/repository"
	"go.mongodb.org/mongo-driver/bson"
)

func TestParseQueryOptions(t *testing.T) {
	type StructField struct {
		NestedField *string `bson:"nested"`
	}
	type TestFilter struct {
		StringField *string `bson:"string"`
		IntField    *int    `bson:"int"`
		BoolField   *bool   `bson:"bool"`
		StructField `bson:"child"`
	}

	testString := "test string"
	testInt := 1
	TestBool := true

	type args struct {
		q *repository.QueryOptions
	}
	tests := []struct {
		name string
		args func() args
		want bson.M
	}{
		{
			name: "empty",
			args: func() args {
				return args{
					q: nil,
				}
			},
			want: bson.M{},
		},
		{
			name: "with fields",
			args: func() args {
				f := TestFilter{
					StringField: &testString,
					IntField:    &testInt,
					BoolField:   &TestBool,
					StructField: StructField{
						&testString,
					},
				}
				return args{
					q: &repository.QueryOptions{
						Filter: f,
					},
				}
			},
			want: bson.M{
				"string":       testString,
				"int":          testInt,
				"bool":         TestBool,
				"child.nested": testString,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQueryOptions(tt.args().q)
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("ParseQueryOptions() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
