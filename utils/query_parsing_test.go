package utils

import (
	"net/url"
	"reflect"
	"testing"
)

func TestFromQuery(t *testing.T) {
	type TestStruct struct {
		SField  string  `schema:"string"`
		IField  int     `schema:"int"`
		BField  bool    `schema:"bool"`
		PSField *string `schema:"pstring"`
		PIField *int    `schema:"pint"`
	}

	pstring := "string"
	pint := 10

	type args struct {
		query url.Values
	}
	tests := []struct {
		name string
		args args
		want *TestStruct
	}{
		{
			name: "empty",
			args: args{},
			want: &TestStruct{},
		},
		{
			name: "with fields",
			args: args{
				query: url.Values{
					"string": []string{"string"},
					"int":    []string{"10"},
					"bool":   []string{"true"},
				},
			},
			want: &TestStruct{
				SField: "string",
				IField: 10,
				BField: true,
			},
		},
		{
			name: "with fields&pointers",
			args: args{
				query: url.Values{
					"string":  []string{"string"},
					"int":     []string{"10"},
					"bool":    []string{"true"},
					"pstring": []string{"string"},
					"pint":    []string{"10"},
				},
			},
			want: &TestStruct{
				SField:  "string",
				IField:  10,
				BField:  true,
				PSField: &pstring,
				PIField: &pint,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromQuery[TestStruct](tt.args.query); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
