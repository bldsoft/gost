package utils

import (
	"context"
	"reflect"
	"testing"
)

func TestProbe(t *testing.T) {
	type tmp struct {
		Duration float64 `json:"duration"`
	}
	type args struct {
		path string
		args map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    tmp
		wantErr bool
	}{
		{
			name: "",
			args: args{
				path: "test_files/media_test.ts",
				args: map[string]interface{}{
					"show_entries": "format=duration",
				},
			},
			want: tmp{
				Duration: 6.013333,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Probe(context.TODO(), tt.args.path, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Probe() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if val, _ := got.Duration(); val != tt.want.Duration {
				t.Errorf("Probe() duration = %v, want.Duration %v", val, tt.want.Duration)
				return
			}
		})
	}
}

func TestProbeInto(t *testing.T) {
	type ffprobeRes struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}

	type args struct {
		path string
		res  *ffprobeRes
		args map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *ffprobeRes
	}{
		{
			name: "success",
			args: args{
				path: "test_files/media_test.ts",
				res:  &ffprobeRes{},
				args: map[string]interface{}{
					"show_entries": "format=duration",
				},
			},
			wantErr: false,
			want: &ffprobeRes{
				Format: struct {
					Duration string "json:\"duration\""
				}{
					Duration: "6.013333",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ProbeInto(context.TODO(), tt.args.path, tt.args.res, tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("ProbeInto() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.res, tt.want) {
				t.Errorf("ProbeInto() got = %v, want %v", tt.args.res, tt.want)
			}
		})
	}
}
